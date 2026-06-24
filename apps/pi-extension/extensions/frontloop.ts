import { StringEnum } from "@earendil-works/pi-ai";
import type { ExtensionAPI, ExtensionContext } from "@earendil-works/pi-coding-agent";
import { Text } from "@earendil-works/pi-tui";
import { Type } from "typebox";
import { constants as fsConstants } from "node:fs";
import { access, appendFile, mkdir, readdir, readFile, rename, stat, writeFile } from "node:fs/promises";
import { basename, dirname, join, relative, resolve } from "node:path";

const STATUSES = ["clarify", "ready", "in_progress", "done"] as const;
const CREATE_STATUSES = ["ready", "clarify"] as const;
const PRIORITIES = ["critical", "high", "medium", "low"] as const;
const DEFAULT_EPIC = "default";
const FRONTLOOP_DIR = ".frontloop";

type Status = (typeof STATUSES)[number];
type CreateStatus = (typeof CREATE_STATUSES)[number];
type Priority = (typeof PRIORITIES)[number];

const STATUS_LABELS: Record<Status, string> = {
	in_progress: "IN PROGRESS",
	ready: "READY",
	clarify: "NEEDS CLARIFICATION",
	done: "DONE",
};

const PRIORITY_ORDER_PREFIX: Record<Priority, string> = {
	critical: "0001",
	high: "2500",
	medium: "5000",
	low: "7500",
};

type TaskSummary = {
	epic: string;
	status: Status;
	filename: string;
	path: string;
	title: string;
	priority: Priority | "unknown";
	goal?: string;
	acceptanceCriteria?: string;
	designDecisions?: string;
	implementationNotes?: string;
	modifiedAt: number;
};

type EpicSnapshot = {
	slug: string;
	tasks: Record<Status, TaskSummary[]>;
};

type QueueSnapshot = {
	root: string;
	epics: EpicSnapshot[];
};

type CreateTaskInput = {
	epic?: string;
	status?: CreateStatus;
	title: string;
	goal: string;
	priority: Priority;
	acceptanceCriteria: string[];
	designDecisions?: string[];
	implementationNotes?: string;
	questions?: string[];
};

type CompleteTaskInput = {
	epic?: string;
	completionSummary: string[];
	filesChanged?: string[];
};

type BlockTaskInput = {
	epic?: string;
	reason: string;
};

let mutationQueue: Promise<void> = Promise.resolve();

async function withFrontloopMutation<T>(fn: () => Promise<T>): Promise<T> {
	const previous = mutationQueue;
	let release!: () => void;
	mutationQueue = new Promise<void>((resolveRelease) => {
		release = resolveRelease;
	});

	await previous;
	try {
		return await fn();
	} finally {
		release();
	}
}

async function exists(path: string): Promise<boolean> {
	try {
		await access(path, fsConstants.F_OK);
		return true;
	} catch {
		return false;
	}
}

async function isDirectory(path: string): Promise<boolean> {
	try {
		return (await stat(path)).isDirectory();
	} catch {
		return false;
	}
}

async function findFrontloopRoot(cwd: string): Promise<string | undefined> {
	let current = resolve(cwd);
	while (true) {
		const candidate = join(current, FRONTLOOP_DIR);
		if (await isV2Root(candidate)) {
			return candidate;
		}

		const parent = dirname(current);
		if (parent === current) {
			return undefined;
		}
		current = parent;
	}
}

async function findLegacyFrontloopRoot(cwd: string): Promise<string | undefined> {
	let current = resolve(cwd);
	while (true) {
		const candidate = join(current, FRONTLOOP_DIR);
		if (await isLegacyRoot(candidate)) {
			return candidate;
		}

		const parent = dirname(current);
		if (parent === current) {
			return undefined;
		}
		current = parent;
	}
}

async function requireFrontloopRoot(cwd: string): Promise<string> {
	const root = await findFrontloopRoot(cwd);
	if (root) {
		return root;
	}

	const legacyRoot = await findLegacyFrontloopRoot(cwd);
	if (legacyRoot) {
		throw new Error(`Legacy flat frontloop queue detected at ${relative(cwd, legacyRoot) || "."}. Run \`fl migrate epic-layout\` from the repository root first.`);
	}

	throw new Error("No frontloop v2 queue found. Expected .frontloop/default/{clarify,ready,in_progress,done}; run `fl init` from the repository root first.");
}

async function isV2Root(root: string): Promise<boolean> {
	const defaultEpic = join(root, DEFAULT_EPIC);
	if (!(await isDirectory(defaultEpic))) {
		return false;
	}

	for (const status of STATUSES) {
		if (!(await isDirectory(join(defaultEpic, status)))) {
			return false;
		}
	}
	return true;
}

async function isLegacyRoot(root: string): Promise<boolean> {
	for (const status of STATUSES) {
		if (!(await isDirectory(join(root, status)))) {
			return false;
		}
	}
	return true;
}

async function listActiveEpics(root: string): Promise<string[]> {
	const entries = await readdir(root, { withFileTypes: true });
	const epics: string[] = [];

	for (const entry of entries) {
		if (!entry.isDirectory() || entry.name.startsWith("_")) {
			continue;
		}
		const epicPath = join(root, entry.name);
		const valid = await Promise.all(STATUSES.map((status) => isDirectory(join(epicPath, status))));
		if (valid.every(Boolean)) {
			epics.push(entry.name);
		}
	}

	return epics.sort((a, b) => {
		if (a === DEFAULT_EPIC) return -1;
		if (b === DEFAULT_EPIC) return 1;
		return a.localeCompare(b);
	});
}

async function requireEpic(root: string, epic = DEFAULT_EPIC): Promise<string> {
	const epics = await listActiveEpics(root);
	if (!epics.includes(epic)) {
		throw new Error(`Unknown active frontloop epic: ${epic}. Active epics: ${epics.join(", ") || "(none)"}`);
	}
	return epic;
}

async function snapshotQueue(root: string, epic?: string): Promise<QueueSnapshot> {
	const selectedEpics = epic ? [await requireEpic(root, epic)] : await listActiveEpics(root);
	const epics: EpicSnapshot[] = [];

	for (const slug of selectedEpics) {
		const tasks = {} as Record<Status, TaskSummary[]>;
		for (const status of STATUSES) {
			tasks[status] = await listTasks(root, slug, status);
		}
		epics.push({ slug, tasks });
	}

	return { root, epics };
}

async function listTasks(root: string, epic: string, status: Status): Promise<TaskSummary[]> {
	const dir = join(root, epic, status);
	const entries = await readdir(dir, { withFileTypes: true });
	const files = entries
		.filter((entry) => entry.isFile() && entry.name.endsWith(".md"))
		.map((entry) => entry.name)
		.sort((a, b) => a.localeCompare(b));

	const tasks: TaskSummary[] = [];
	for (const filename of files) {
		const path = join(dir, filename);
		tasks.push(await readTask(path, epic, status));
	}
	return tasks;
}

async function readTask(path: string, epic: string, status: Status): Promise<TaskSummary> {
	const content = await readFile(path, "utf8");
	const metadata = parseTaskMarkdown(content);
	const stats = await stat(path);
	return {
		epic,
		status,
		filename: basename(path),
		path,
		title: metadata.title || stripMarkdownExtension(stripNumericPrefix(basename(path))),
		priority: normalizePriority(metadata.priority),
		goal: metadata.goal,
		acceptanceCriteria: metadata.acceptanceCriteria,
		designDecisions: metadata.designDecisions,
		implementationNotes: metadata.implementationNotes,
		modifiedAt: stats.mtimeMs,
	};
}

function parseTaskMarkdown(content: string): {
	title?: string;
	priority?: string;
	goal?: string;
	acceptanceCriteria?: string;
	designDecisions?: string;
	implementationNotes?: string;
} {
	const frontmatter = content.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n?([\s\S]*)$/);
	const body = frontmatter ? frontmatter[2] : content;
	const fields: Record<string, string> = {};

	if (frontmatter) {
		for (const line of frontmatter[1].split(/\r?\n/)) {
			const match = line.match(/^([A-Za-z0-9_-]+):\s*(.*?)\s*$/);
			if (!match) continue;
			fields[match[1]] = unquoteYamlScalar(match[2]);
		}
	}

	return {
		title: fields.title,
		priority: fields.priority,
		goal: extractSection(body, "Goal"),
		acceptanceCriteria: extractSection(body, "Acceptance Criteria"),
		designDecisions: extractSection(body, "Design Decisions"),
		implementationNotes: extractSection(body, "Implementation Notes"),
	};
}

function extractSection(body: string, heading: string): string | undefined {
	const escaped = heading.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
	const headingRegex = new RegExp(`^##\\s+${escaped}\\s*$`, "mi");
	const match = headingRegex.exec(body);
	if (!match) return undefined;

	const afterHeading = match.index + match[0].length;
	const remainder = body.slice(afterHeading).replace(/^\r?\n/, "");
	const nextHeading = /^##\s+/m.exec(remainder);
	const section = nextHeading ? remainder.slice(0, nextHeading.index) : remainder;
	const trimmed = section.trim();
	return trimmed.length > 0 ? trimmed : undefined;
}

function unquoteYamlScalar(value: string): string {
	const trimmed = value.trim();
	if ((trimmed.startsWith('"') && trimmed.endsWith('"')) || (trimmed.startsWith("'") && trimmed.endsWith("'"))) {
		return trimmed.slice(1, -1);
	}
	return trimmed;
}

function normalizePriority(value: string | undefined): Priority | "unknown" {
	return PRIORITIES.includes(value as Priority) ? (value as Priority) : "unknown";
}

function renderStatus(snapshot: QueueSnapshot, cwd: string): string {
	const lines: string[] = ["=== Frontloop Status ===", ""];

	if (snapshot.epics.length === 0) {
		lines.push("No active epics found.");
		return lines.join("\n");
	}

	for (const epic of snapshot.epics) {
		lines.push(`EPIC: ${epic.slug}${epic.slug === DEFAULT_EPIC ? " (default)" : ""}`);
		lines.push("");

		for (const status of ["in_progress", "ready", "clarify", "done"] as Status[]) {
			const tasks = status === "done" ? sortRecent(epic.tasks[status]).slice(0, 5) : epic.tasks[status];
			const total = epic.tasks[status].length;
			lines.push(`${STATUS_LABELS[status]} (${total}):`);
			if (tasks.length === 0) {
				lines.push("  (empty)");
			} else {
				for (const task of tasks) {
					lines.push(`  ${stripMarkdownExtension(task.filename)}  [${task.priority}]  ${task.title}`);
				}
				if (status === "done" && total > tasks.length) {
					lines.push(`  ... and ${total - tasks.length} more`);
				}
			}
			lines.push("");
		}
	}

	lines.push(`Root: ${relativeOrSelf(cwd, snapshot.root)}`);
	return lines.join("\n").trimEnd();
}

function sortRecent(tasks: TaskSummary[]): TaskSummary[] {
	return [...tasks].sort((a, b) => b.modifiedAt - a.modifiedAt);
}

async function createTask(cwd: string, input: CreateTaskInput): Promise<TaskSummary> {
	return withFrontloopMutation(async () => {
		const root = await requireFrontloopRoot(cwd);
		const epic = await requireEpic(root, input.epic || DEFAULT_EPIC);
		const questionCount = input.questions?.filter(Boolean).length || 0;
		const status = normalizeCreateStatus(input.status, questionCount > 0 ? "clarify" : "ready");
		if (status === "ready" && questionCount > 0) {
			throw new Error("Ready frontloop tasks cannot include open Questions. Create it in clarify or remove the questions.");
		}

		const dir = join(root, epic, status);
		const baseName = `${slugify(input.title)}.md`;
		const preferred = status === "ready" ? `${PRIORITY_ORDER_PREFIX[input.priority]}-${baseName}` : baseName;
		const filename = await uniqueFilename(dir, preferred);
		const path = join(dir, filename);
		const content = formatNewTask(input);

		await writeFile(path, content, { encoding: "utf8", flag: "wx" });
		return readTask(path, epic, status);
	});
}

function normalizeCreateStatus(value: string | undefined, defaultStatus: CreateStatus): CreateStatus {
	if (!value) return defaultStatus;
	if ((CREATE_STATUSES as readonly string[]).includes(value)) return value as CreateStatus;
	throw new Error(`Unsupported task creation status: ${value}`);
}

function formatNewTask(input: CreateTaskInput): string {
	const lines: string[] = [
		"---",
		`title: ${sanitizeInline(input.title)}`,
		`priority: ${input.priority}`,
		"---",
		"",
		"## Goal",
		"",
		input.goal.trim(),
		"",
		"## Acceptance Criteria",
		"",
		...input.acceptanceCriteria.filter(Boolean).map((criterion) => `- ${criterion.trim()}`),
	];

	if (input.designDecisions?.filter(Boolean).length) {
		lines.push("", "## Design Decisions", "", ...input.designDecisions.filter(Boolean).map((decision) => `- ${decision.trim()}`));
	}

	if (input.implementationNotes?.trim()) {
		lines.push("", "## Implementation Notes", "", input.implementationNotes.trim());
	}

	if (input.questions?.filter(Boolean).length) {
		lines.push("", "## Questions", "");
		input.questions.filter(Boolean).forEach((question, index) => {
			const trimmed = question.trim();
			if (/^###\s+Q\d+:/i.test(trimmed)) {
				lines.push(trimmed);
			} else {
				lines.push(`### Q${index + 1}: ${trimmed}`);
			}
			lines.push("");
		});
	}

	return `${lines.join("\n").replace(/\n{3,}/g, "\n\n").trim()}\n`;
}

async function startNextTask(cwd: string, requestedEpic?: string): Promise<TaskSummary> {
	return withFrontloopMutation(async () => {
		const root = await requireFrontloopRoot(cwd);
		const epic = await chooseReadyEpic(root, requestedEpic);
		const inProgress = await listTasks(root, epic, "in_progress");
		if (inProgress.length > 0) {
			throw new Error(`Epic ${epic} already has in-progress task(s): ${inProgress.map((task) => stripMarkdownExtension(task.filename)).join(", ")}`);
		}

		const ready = await listTasks(root, epic, "ready");
		if (ready.length === 0) {
			throw new Error(`No ready tasks in epic ${epic}.`);
		}

		const task = ready[0];
		const target = join(root, epic, "in_progress", task.filename);
		await rename(task.path, target);
		return readTask(target, epic, "in_progress");
	});
}

async function chooseReadyEpic(root: string, requestedEpic?: string): Promise<string> {
	if (requestedEpic) {
		return requireEpic(root, requestedEpic);
	}

	const epics = await listActiveEpics(root);
	const readyEpics: string[] = [];
	for (const epic of epics) {
		if ((await listTasks(root, epic, "ready")).length > 0) {
			readyEpics.push(epic);
		}
	}

	if (readyEpics.length === 0) {
		throw new Error("No tasks ready for work.");
	}
	if (readyEpics.length > 1) {
		throw new Error(`Multiple epics have ready tasks. Specify one of: ${readyEpics.join(", ")}`);
	}
	return readyEpics[0];
}

async function completeTask(cwd: string, input: CompleteTaskInput): Promise<TaskSummary> {
	return withFrontloopMutation(async () => {
		const root = await requireFrontloopRoot(cwd);
		const task = await selectInProgressTask(root, input.epic);
		const completion = formatCompletionSummary(input.completionSummary, input.filesChanged || []);
		await appendFile(task.path, completion, "utf8");

		const doneName = await uniqueFilename(join(root, task.epic, "done"), task.filename);
		const target = join(root, task.epic, "done", doneName);
		await rename(task.path, target);
		return readTask(target, task.epic, "done");
	});
}

async function blockTask(cwd: string, input: BlockTaskInput): Promise<TaskSummary> {
	return withFrontloopMutation(async () => {
		const root = await requireFrontloopRoot(cwd);
		const task = await selectInProgressTask(root, input.epic);
		await appendFile(task.path, `\n\n## Blocked\n\n${input.reason.trim()}\n`, "utf8");

		const clarifyName = await uniqueFilename(join(root, task.epic, "clarify"), stripNumericPrefix(task.filename));
		const target = join(root, task.epic, "clarify", clarifyName);
		await rename(task.path, target);
		return readTask(target, task.epic, "clarify");
	});
}

async function selectInProgressTask(root: string, requestedEpic?: string): Promise<TaskSummary> {
	const epics = requestedEpic ? [await requireEpic(root, requestedEpic)] : await listActiveEpics(root);
	const tasks: TaskSummary[] = [];
	for (const epic of epics) {
		tasks.push(...(await listTasks(root, epic, "in_progress")));
	}

	if (tasks.length === 0) {
		throw new Error(requestedEpic ? `No in-progress task in epic ${requestedEpic}.` : "No in-progress frontloop task found.");
	}
	if (tasks.length > 1) {
		throw new Error(`Multiple in-progress tasks found. Specify an epic: ${tasks.map((task) => task.epic).join(", ")}`);
	}
	return tasks[0];
}

function formatCompletionSummary(summary: string[], filesChanged: string[]): string {
	const lines = ["", "", "## Completion Summary", ""];
	for (const item of summary.filter(Boolean)) {
		lines.push(`- ${item.trim()}`);
	}

	if (filesChanged.filter(Boolean).length) {
		lines.push("", "### Files Changed", "");
		for (const file of filesChanged.filter(Boolean)) {
			lines.push(`- ${file.trim()}`);
		}
	}

	return `${lines.join("\n")}\n`;
}

function buildWorkPrompt(task: TaskSummary, cwd: string): string {
	const path = relativeOrSelf(cwd, task.path);
	const sections = [
		`Work the active frontloop task now.`,
		`Task file: ${path}`,
		`Epic: ${task.epic}`,
		`Title: ${task.title}`,
		`Priority: ${task.priority}`,
	];

	if (task.goal) sections.push(`\n## Goal\n\n${task.goal}`);
	if (task.acceptanceCriteria) sections.push(`\n## Acceptance Criteria\n\n${task.acceptanceCriteria}`);
	if (task.designDecisions) sections.push(`\n## Design Decisions\n\n${task.designDecisions}`);
	if (task.implementationNotes) sections.push(`\n## Implementation Notes\n\n${task.implementationNotes}`);

	sections.push(
		"\n## Frontloop instructions\n\n" +
			"- Follow the task's acceptance criteria exactly.\n" +
			"- Do not modify the task's Goal, Acceptance Criteria, or Design Decisions sections.\n" +
			"- When complete, call frontloop_complete_task with concise summary bullets and changed files.\n" +
			"- If blocked, call frontloop_block_task with the blocker reason.",
	);

	return sections.join("\n");
}

async function activeTaskContext(cwd: string): Promise<string | undefined> {
	const root = await findFrontloopRoot(cwd);
	if (!root) return undefined;

	const epics = await listActiveEpics(root);
	const tasks: TaskSummary[] = [];
	for (const epic of epics) {
		tasks.push(...(await listTasks(root, epic, "in_progress")));
	}
	if (tasks.length === 0) return undefined;

	return [
		"\n\nFrontloop active task context:",
		...tasks.map((task) => {
			const lines = [
				`- ${task.epic}/${stripMarkdownExtension(task.filename)}: ${task.title} [${task.priority}]`,
				task.goal ? `  Goal: ${collapseWhitespace(task.goal)}` : undefined,
				`  Task file: ${relativeOrSelf(cwd, task.path)}`,
			];
			return lines.filter(Boolean).join("\n");
		}),
		"Use frontloop_complete_task when done or frontloop_block_task if blocked.",
	].join("\n");
}

async function updateStatus(ctx: ExtensionContext): Promise<void> {
	if (!ctx.hasUI) return;
	const root = await findFrontloopRoot(ctx.cwd);
	if (!root) {
		ctx.ui.setStatus("frontloop", undefined);
		return;
	}

	const snapshot = await snapshotQueue(root);
	let ready = 0;
	let clarify = 0;
	let inProgress = 0;
	for (const epic of snapshot.epics) {
		ready += epic.tasks.ready.length;
		clarify += epic.tasks.clarify.length;
		inProgress += epic.tasks.in_progress.length;
	}

	const theme = ctx.ui.theme;
	const bits = [`${ready} ready`, `${clarify} clarify`];
	if (inProgress > 0) bits.unshift(`${inProgress} active`);
	ctx.ui.setStatus("frontloop", theme.fg(inProgress > 0 ? "accent" : "dim", `frontloop: ${bits.join(", ")}`));
}

function stripNumericPrefix(filename: string): string {
	return filename.replace(/^\d{4}-/, "");
}

function stripMarkdownExtension(filename: string): string {
	return filename.replace(/\.md$/i, "");
}

function slugify(title: string): string {
	const slug = title
		.toLowerCase()
		.replace(/['’]/g, "")
		.replace(/[^a-z0-9]+/g, "-")
		.replace(/^-+|-+$/g, "");
	return slug || "task";
}

async function uniqueFilename(dir: string, preferred: string): Promise<string> {
	await mkdir(dir, { recursive: true });
	const ext = preferred.endsWith(".md") ? ".md" : "";
	const base = ext ? preferred.slice(0, -ext.length) : preferred;
	let candidate = `${base}${ext}`;
	let index = 2;
	while (await exists(join(dir, candidate))) {
		candidate = `${base}-${index}${ext}`;
		index++;
	}
	return candidate;
}

function sanitizeInline(value: string): string {
	return value.replace(/[\r\n]+/g, " ").trim();
}

function collapseWhitespace(value: string): string {
	return value.replace(/\s+/g, " ").trim();
}

function relativeOrSelf(from: string, target: string): string {
	const rel = relative(from, target);
	return rel && !rel.startsWith("..") ? rel : target;
}

function parseLines(value: string | undefined): string[] {
	return (value || "")
		.split(/\r?\n/)
		.map((line) => line.trim().replace(/^-\s+/, ""))
		.filter(Boolean);
}

async function showFrontloopMessage(pi: ExtensionAPI, ctx: ExtensionContext, content: string, details?: Record<string, unknown>) {
	pi.sendMessage({ customType: "frontloop", content, display: true, details });
	await updateStatus(ctx);
}

export default function frontloopExtension(pi: ExtensionAPI) {
	pi.registerMessageRenderer("frontloop", (message, _options, theme) => {
		const content = typeof message.content === "string" ? message.content : JSON.stringify(message.content, null, 2);
		return new Text(theme.fg("accent", "frontloop") + "\n" + content, 0, 0);
	});

	pi.on("session_start", async (_event, ctx) => {
		await updateStatus(ctx);
	});

	pi.on("agent_end", async (_event, ctx) => {
		await updateStatus(ctx);
	});

	pi.on("before_agent_start", async (event, ctx) => {
		const context = await activeTaskContext(ctx.cwd);
		if (!context) return undefined;
		return { systemPrompt: `${event.systemPrompt}${context}` };
	});

	pi.registerCommand("fl-work", {
		description: "Start the next ready frontloop task and send it to the agent",
		handler: async (args, ctx) => {
			try {
				const root = await requireFrontloopRoot(ctx.cwd);
				let epic = args.trim() || undefined;
				if (!epic) {
					const epics = await listActiveEpics(root);
					const activeTasks: TaskSummary[] = [];
					const readyEpics: string[] = [];
					for (const candidate of epics) {
						activeTasks.push(...(await listTasks(root, candidate, "in_progress")));
						if ((await listTasks(root, candidate, "ready")).length > 0) readyEpics.push(candidate);
					}
					if (activeTasks.length === 1) {
						pi.sendUserMessage(buildWorkPrompt(activeTasks[0], ctx.cwd));
						await updateStatus(ctx);
						return;
					}
					if (readyEpics.length > 1) {
						if (!ctx.hasUI) throw new Error(`Multiple epics have ready tasks. Specify one of: ${readyEpics.join(", ")}`);
						epic = await ctx.ui.select("Which epic should /fl-work use?", readyEpics);
						if (!epic) return;
					}
				}

				const existing = epic ? await listTasks(root, await requireEpic(root, epic), "in_progress") : [];
				if (existing.length > 0) {
					const task = existing[0];
					pi.sendUserMessage(buildWorkPrompt(task, ctx.cwd));
					await updateStatus(ctx);
					return;
				}

				const task = await startNextTask(ctx.cwd, epic);
				await updateStatus(ctx);
				pi.sendUserMessage(buildWorkPrompt(task, ctx.cwd));
			} catch (error) {
				ctx.ui.notify(error instanceof Error ? error.message : String(error), "error");
			}
		},
	});

	pi.registerCommand("fl-status", {
		description: "Show the active frontloop queue grouped by epic",
		handler: async (args, ctx) => {
			try {
				const root = await requireFrontloopRoot(ctx.cwd);
				const epic = args.trim() || undefined;
				const snapshot = await snapshotQueue(root, epic);
				await showFrontloopMessage(pi, ctx, renderStatus(snapshot, ctx.cwd), { kind: "status", epic });
			} catch (error) {
				ctx.ui.notify(error instanceof Error ? error.message : String(error), "error");
			}
		},
	});

	pi.registerCommand("fl-add", {
		description: "Create a frontloop task in ready or clarify",
		handler: async (args, ctx) => {
			try {
				if (!ctx.hasUI) throw new Error("/fl-add requires interactive UI.");
				const root = await requireFrontloopRoot(ctx.cwd);
				const epics = await listActiveEpics(root);
				const selectedEpic = await ctx.ui.select("Target epic", epics);
				if (!selectedEpic) return;

				const title = (args.trim() || (await ctx.ui.input("Task title", "Short human-readable title")) || "").trim();
				if (!title) return;
				const priority = (await ctx.ui.select("Priority", [...PRIORITIES])) as Priority | undefined;
				if (!priority) return;
				const goal = (await ctx.ui.editor("Goal", "What should this task achieve?"))?.trim();
				if (!goal) return;
				const criteria = parseLines(await ctx.ui.editor("Acceptance criteria", "One criterion per line"));
				if (criteria.length === 0) return;
				const decisions = parseLines(await ctx.ui.editor("Design decisions (optional)", "One decision per line"));
				const notes = (await ctx.ui.editor("Implementation notes (optional)", "Relevant files, constraints, hints"))?.trim();
				const status = (await ctx.ui.select("Initial status", [...CREATE_STATUSES])) as CreateStatus | undefined;
				if (!status) return;

				const task = await createTask(ctx.cwd, {
					epic: selectedEpic,
					status,
					title,
					priority,
					goal,
					acceptanceCriteria: criteria,
					designDecisions: decisions,
					implementationNotes: notes,
				});
				await showFrontloopMessage(pi, ctx, `Created ${relativeOrSelf(ctx.cwd, task.path)}`, { kind: "created", task });
			} catch (error) {
				ctx.ui.notify(error instanceof Error ? error.message : String(error), "error");
			}
		},
	});

	pi.registerCommand("fl-complete", {
		description: "Complete the active frontloop task and move it to done",
		handler: async (args, ctx) => {
			try {
				if (!ctx.hasUI) throw new Error("/fl-complete requires interactive UI.");
				const completionSummary = parseLines(await ctx.ui.editor("Completion summary", "One bullet per completed change"));
				if (completionSummary.length === 0) return;
				const filesChanged = parseLines(await ctx.ui.editor("Files changed", "One path per line"));
				const task = await completeTask(ctx.cwd, { epic: args.trim() || undefined, completionSummary, filesChanged });
				await showFrontloopMessage(pi, ctx, `Completed ${relativeOrSelf(ctx.cwd, task.path)}`, { kind: "completed", task });
			} catch (error) {
				ctx.ui.notify(error instanceof Error ? error.message : String(error), "error");
			}
		},
	});

	pi.registerCommand("fl-block", {
		description: "Mark the active frontloop task blocked and return it to clarify",
		handler: async (args, ctx) => {
			try {
				if (!ctx.hasUI) throw new Error("/fl-block requires interactive UI.");
				const reason = (await ctx.ui.editor("Blocker reason", "What prevents this task from being completed?"))?.trim();
				if (!reason) return;
				const task = await blockTask(ctx.cwd, { epic: args.trim() || undefined, reason });
				await showFrontloopMessage(pi, ctx, `Blocked ${relativeOrSelf(ctx.cwd, task.path)}`, { kind: "blocked", task });
			} catch (error) {
				ctx.ui.notify(error instanceof Error ? error.message : String(error), "error");
			}
		},
	});

	const PrioritySchema = StringEnum(PRIORITIES);
	const CreateStatusSchema = StringEnum(CREATE_STATUSES);

	pi.registerTool({
		name: "frontloop_status",
		label: "Frontloop Status",
		description: "Show the active frontloop v2 queue grouped by epic.",
		promptSnippet: "Inspect the active frontloop v2 queue grouped by epic",
		promptGuidelines: [
			"Use frontloop_status to inspect frontloop queues instead of manually listing .frontloop directories.",
			"Frontloop extension tools expect the v2 .frontloop/default layout and do not perform legacy queue migration.",
		],
		parameters: Type.Object({
			epic: Type.Optional(Type.String({ description: "Optional active epic slug to inspect" })),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, ctx) {
			const root = await requireFrontloopRoot(ctx.cwd);
			const snapshot = await snapshotQueue(root, params.epic);
			const text = renderStatus(snapshot, ctx.cwd);
			await updateStatus(ctx);
			return { content: [{ type: "text", text }], details: snapshot };
		},
	});

	pi.registerTool({
		name: "frontloop_create_task",
		label: "Frontloop Create Task",
		description: "Create a frontloop task in .frontloop/<epic>/ready/ when actionable, or clarify/ when open questions remain.",
		promptSnippet: "Create a frontloop task in an epic's ready or clarify queue",
		promptGuidelines: [
			"Use frontloop_create_task when the user asks to capture work as a frontloop task.",
			"Create directly in ready when the task is actionable and has no open questions, especially when the user wants to work on it next.",
			"Use clarify only when human decisions or missing details remain; ready tasks must not include a Questions section.",
		],
		parameters: Type.Object({
			epic: Type.Optional(Type.String({ description: "Active epic slug; defaults to default" })),
			status: Type.Optional(CreateStatusSchema),
			title: Type.String({ description: "Short human-readable task title" }),
			goal: Type.String({ description: "What the task achieves, in 1-3 sentences" }),
			priority: PrioritySchema,
			acceptanceCriteria: Type.Array(Type.String(), { description: "Concrete checklist items" }),
			designDecisions: Type.Optional(Type.Array(Type.String(), { description: "Pre-approved design choices" })),
			implementationNotes: Type.Optional(Type.String({ description: "Optional hints, constraints, or relevant files" })),
			questions: Type.Optional(Type.Array(Type.String(), { description: "Optional clarify questions as markdown, ideally with options and recommendation. Supplying questions defaults the task to clarify." })),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, ctx) {
			const task = await createTask(ctx.cwd, params as CreateTaskInput);
			await updateStatus(ctx);
			return {
				content: [{ type: "text", text: `Created ${relativeOrSelf(ctx.cwd, task.path)}` }],
				details: { task },
			};
		},
	});

	pi.registerTool({
		name: "frontloop_start_task",
		label: "Frontloop Start Task",
		description: "Move the next ready task in an epic to in_progress and return its task context.",
		promptSnippet: "Start the next ready frontloop task by moving it to in_progress",
		promptGuidelines: [
			"Use frontloop_start_task before executing a ready frontloop task; it preserves epic membership and task ordering.",
		],
		parameters: Type.Object({
			epic: Type.Optional(Type.String({ description: "Active epic slug. Required if multiple epics have ready tasks." })),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, ctx) {
			const task = await startNextTask(ctx.cwd, params.epic);
			await updateStatus(ctx);
			return {
				content: [{ type: "text", text: buildWorkPrompt(task, ctx.cwd) }],
				details: { task },
			};
		},
	});

	pi.registerTool({
		name: "frontloop_complete_task",
		label: "Frontloop Complete Task",
		description: "Append a completion summary to the active in_progress task and move it to done.",
		promptSnippet: "Complete the active frontloop task and move it to done",
		promptGuidelines: [
			"Use frontloop_complete_task after satisfying a frontloop task's acceptance criteria.",
			"frontloop_complete_task should include concise summary bullets and changed file paths.",
		],
		parameters: Type.Object({
			epic: Type.Optional(Type.String({ description: "Active epic slug. Required if multiple epics have in-progress tasks." })),
			completionSummary: Type.Array(Type.String(), { description: "One bullet per completed change" }),
			filesChanged: Type.Optional(Type.Array(Type.String(), { description: "Changed file paths, optionally with status notes" })),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, ctx) {
			const task = await completeTask(ctx.cwd, params as CompleteTaskInput);
			await updateStatus(ctx);
			return {
				content: [{ type: "text", text: `Completed ${relativeOrSelf(ctx.cwd, task.path)}` }],
				details: { task },
			};
		},
	});

	pi.registerTool({
		name: "frontloop_block_task",
		label: "Frontloop Block Task",
		description: "Append a blocker reason to the active in_progress task and move it back to clarify without its numeric prefix.",
		promptSnippet: "Block the active frontloop task and return it to clarify",
		promptGuidelines: [
			"Use frontloop_block_task when a frontloop task cannot be completed as described and needs human clarification.",
		],
		parameters: Type.Object({
			epic: Type.Optional(Type.String({ description: "Active epic slug. Required if multiple epics have in-progress tasks." })),
			reason: Type.String({ description: "Why the task is blocked" }),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, ctx) {
			const task = await blockTask(ctx.cwd, params as BlockTaskInput);
			await updateStatus(ctx);
			return {
				content: [{ type: "text", text: `Blocked ${relativeOrSelf(ctx.cwd, task.path)}` }],
				details: { task },
			};
		},
	});
}
