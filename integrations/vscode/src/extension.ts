import * as vscode from 'vscode';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

let outputChannel: vscode.OutputChannel;
let statusBarItem: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
    console.log('Murmur AI extension activated');

    // Create output channel
    outputChannel = vscode.window.createOutputChannel('Murmur');

    // Create status bar item
    statusBarItem = vscode.window.createStatusBarItem(
        vscode.StatusBarAlignment.Left,
        100
    );
    statusBarItem.command = 'murmur.patterns';
    statusBarItem.tooltip = 'Click to view patterns';
    context.subscriptions.push(statusBarItem);

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('murmur.sync', cmdSync),
        vscode.commands.registerCommand('murmur.learn.extract', cmdLearnExtract),
        vscode.commands.registerCommand('murmur.stats', cmdStats),
        vscode.commands.registerCommand('murmur.patterns', cmdPatterns)
    );

    // Update status bar on activation
    updateStatusBar();
}

export function deactivate() {
    if (outputChannel) {
        outputChannel.dispose();
    }
    if (statusBarItem) {
        statusBarItem.dispose();
    }
}

// Execute mur CLI command
async function runMur(args: string): Promise<string> {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
    const cwd = workspaceFolder || process.env.HOME || '/';

    try {
        const { stdout, stderr } = await execAsync(`mur ${args}`, { cwd });
        if (stderr) {
            console.warn('mur stderr:', stderr);
        }
        return stdout;
    } catch (error: any) {
        const message = error.stderr || error.message || 'Unknown error';
        throw new Error(`mur ${args} failed: ${message}`);
    }
}

// Command: Sync All
async function cmdSync() {
    outputChannel.show();
    outputChannel.appendLine('--- Murmur Sync All ---');
    outputChannel.appendLine(`[${new Date().toLocaleTimeString()}] Starting sync...`);

    try {
        const result = await runMur('sync all');
        outputChannel.appendLine(result);
        outputChannel.appendLine('✅ Sync completed');
        vscode.window.showInformationMessage('Murmur: Sync completed');
        updateStatusBar();
    } catch (error: any) {
        outputChannel.appendLine(`❌ Error: ${error.message}`);
        vscode.window.showErrorMessage(`Murmur sync failed: ${error.message}`);
    }
}

// Command: Extract Learnings
async function cmdLearnExtract() {
    outputChannel.show();
    outputChannel.appendLine('--- Murmur Extract Learnings ---');
    outputChannel.appendLine(`[${new Date().toLocaleTimeString()}] Extracting...`);

    try {
        const result = await runMur('learn extract --auto');
        outputChannel.appendLine(result);
        outputChannel.appendLine('✅ Extraction completed');
        vscode.window.showInformationMessage('Murmur: Learning extraction completed');
        updateStatusBar();
    } catch (error: any) {
        outputChannel.appendLine(`❌ Error: ${error.message}`);
        vscode.window.showErrorMessage(`Murmur extract failed: ${error.message}`);
    }
}

// Command: Show Stats
async function cmdStats() {
    outputChannel.show();
    outputChannel.appendLine('--- Murmur Stats ---');
    outputChannel.appendLine(`[${new Date().toLocaleTimeString()}]`);

    try {
        const result = await runMur('stats --json');

        // Try to parse and format JSON
        try {
            const stats = JSON.parse(result);
            outputChannel.appendLine(JSON.stringify(stats, null, 2));
        } catch {
            // If not valid JSON, show raw output
            outputChannel.appendLine(result);
        }
    } catch (error: any) {
        outputChannel.appendLine(`❌ Error: ${error.message}`);
        vscode.window.showErrorMessage(`Murmur stats failed: ${error.message}`);
    }
}

// Command: List Patterns
async function cmdPatterns() {
    outputChannel.show();
    outputChannel.appendLine('--- Murmur Patterns ---');
    outputChannel.appendLine(`[${new Date().toLocaleTimeString()}]`);

    try {
        const result = await runMur('pattern list');
        outputChannel.appendLine(result || 'No patterns found');
    } catch (error: any) {
        outputChannel.appendLine(`❌ Error: ${error.message}`);
        vscode.window.showErrorMessage(`Murmur patterns failed: ${error.message}`);
    }
}

// Update status bar with pattern count
async function updateStatusBar() {
    try {
        const result = await runMur('stats --json');
        const stats = JSON.parse(result);
        const count = stats.patterns_count ?? stats.patterns ?? 0;
        statusBarItem.text = `📝 ${count} patterns`;
        statusBarItem.show();
    } catch {
        // If stats fail, try pattern list and count lines
        try {
            const result = await runMur('pattern list');
            const lines = result.trim().split('\n').filter(l => l.trim());
            statusBarItem.text = `📝 ${lines.length} patterns`;
            statusBarItem.show();
        } catch {
            statusBarItem.text = '📝 murmur';
            statusBarItem.show();
        }
    }
}
