"""
MurmurAI - Sublime Text Plugin

Learn and share coding patterns with AI agents.
Requires: mur CLI (https://github.com/anthropics/murmur-ai)
"""

import sublime
import sublime_plugin
import subprocess
import threading
import os


def get_output_panel(window, name="murmur"):
    """Get or create an output panel."""
    panel = window.find_output_panel(name)
    if panel is None:
        panel = window.create_output_panel(name)
    panel.settings().set("word_wrap", True)
    panel.settings().set("line_numbers", False)
    panel.settings().set("gutter", False)
    return panel


def show_output(window, content, name="murmur"):
    """Display content in the output panel."""
    panel = get_output_panel(window, name)
    panel.run_command("append", {"characters": content + "\n"})
    window.run_command("show_panel", {"panel": f"output.{name}"})


def clear_output(window, name="murmur"):
    """Clear the output panel."""
    panel = get_output_panel(window, name)
    panel.run_command("select_all")
    panel.run_command("right_delete")


def run_mur_command(window, args, status_msg, clear=True):
    """Run a mur command asynchronously and show output."""
    
    def run():
        if clear:
            sublime.set_timeout(lambda: clear_output(window), 0)
        
        sublime.set_timeout(
            lambda: show_output(window, f"$ mur {' '.join(args)}\n"), 0
        )
        
        try:
            # Get working directory from active view or use home
            view = window.active_view()
            cwd = None
            if view and view.file_name():
                cwd = os.path.dirname(view.file_name())
            
            result = subprocess.run(
                ["mur"] + args,
                capture_output=True,
                text=True,
                cwd=cwd
            )
            
            output = result.stdout
            if result.stderr:
                output += "\n" + result.stderr
            
            if output.strip():
                sublime.set_timeout(lambda: show_output(window, output), 0)
            
            if result.returncode == 0:
                sublime.set_timeout(
                    lambda: sublime.status_message(f"Murmur: {status_msg} complete"), 0
                )
            else:
                sublime.set_timeout(
                    lambda: sublime.status_message(f"Murmur: {status_msg} failed"), 0
                )
                
        except FileNotFoundError:
            error_msg = (
                "Error: 'mur' command not found.\n"
                "Please install murmur-ai CLI and ensure it's in your PATH.\n"
                "See: https://github.com/anthropics/murmur-ai"
            )
            sublime.set_timeout(lambda: show_output(window, error_msg), 0)
            sublime.set_timeout(
                lambda: sublime.status_message("Murmur: CLI not found"), 0
            )
        except Exception as e:
            sublime.set_timeout(
                lambda: show_output(window, f"Error: {str(e)}"), 0
            )
            sublime.set_timeout(
                lambda: sublime.status_message("Murmur: Error occurred"), 0
            )
    
    sublime.status_message(f"Murmur: {status_msg}...")
    thread = threading.Thread(target=run)
    thread.start()


class MurmurSyncCommand(sublime_plugin.WindowCommand):
    """Run mur sync all."""
    
    def run(self):
        run_mur_command(self.window, ["sync", "all"], "Syncing")


class MurmurLearnExtractCommand(sublime_plugin.WindowCommand):
    """Run mur learn extract --auto."""
    
    def run(self):
        run_mur_command(self.window, ["learn", "extract", "--auto"], "Extracting learnings")


class MurmurStatsCommand(sublime_plugin.WindowCommand):
    """Show murmur statistics."""
    
    def run(self):
        run_mur_command(self.window, ["stats"], "Loading stats")


class MurmurPatternsCommand(sublime_plugin.WindowCommand):
    """List all patterns."""
    
    def run(self):
        run_mur_command(self.window, ["patterns", "list"], "Loading patterns")


class MurmurOpenSettingsCommand(sublime_plugin.WindowCommand):
    """Open MurmurAI settings."""
    
    def run(self):
        self.window.run_command(
            "open_file",
            {"file": "${packages}/MurmurAI/MurmurAI.sublime-settings"}
        )
