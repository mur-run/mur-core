package com.murmurai.plugin

import com.intellij.notification.NotificationGroupManager
import com.intellij.notification.NotificationType
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.progress.ProgressIndicator
import com.intellij.openapi.progress.ProgressManager
import com.intellij.openapi.progress.Task
import com.intellij.openapi.project.Project
import java.io.BufferedReader
import java.io.InputStreamReader
import java.util.concurrent.TimeUnit

/**
 * Base class for Murmur AI actions.
 * Executes `mur` CLI commands and displays results via notifications.
 */
abstract class MurmurAction(
    private val actionName: String,
    private val command: List<String>
) : AnAction() {

    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        
        ProgressManager.getInstance().run(object : Task.Backgroundable(
            project,
            "Murmur AI: $actionName",
            true
        ) {
            override fun run(indicator: ProgressIndicator) {
                indicator.isIndeterminate = true
                indicator.text = "Running: ${command.joinToString(" ")}"
                
                try {
                    val result = executeCommand(project)
                    ApplicationManager.getApplication().invokeLater {
                        showNotification(project, result.first, result.second)
                    }
                } catch (ex: Exception) {
                    ApplicationManager.getApplication().invokeLater {
                        showNotification(
                            project,
                            "Error: ${ex.message ?: "Unknown error"}",
                            NotificationType.ERROR
                        )
                    }
                }
            }
        })
    }

    private fun executeCommand(project: Project): Pair<String, NotificationType> {
        val workDir = project.basePath?.let { java.io.File(it) }
        
        val processBuilder = ProcessBuilder(command)
            .apply {
                workDir?.let { directory(it) }
                redirectErrorStream(true)
            }
        
        val process = processBuilder.start()
        val output = StringBuilder()
        
        BufferedReader(InputStreamReader(process.inputStream)).use { reader ->
            var line: String?
            while (reader.readLine().also { line = it } != null) {
                output.appendLine(line)
            }
        }
        
        val exitCode = process.waitFor(60, TimeUnit.SECONDS)
        val exitValue = if (exitCode) process.exitValue() else -1
        
        return if (exitValue == 0) {
            Pair(output.toString().ifBlank { "$actionName completed successfully" }, NotificationType.INFORMATION)
        } else {
            Pair("Command failed (exit code: $exitValue)\n$output", NotificationType.WARNING)
        }
    }

    private fun showNotification(project: Project, content: String, type: NotificationType) {
        NotificationGroupManager.getInstance()
            .getNotificationGroup("Murmur AI Notifications")
            .createNotification(
                "Murmur AI: $actionName",
                content.take(500), // Truncate long output
                type
            )
            .notify(project)
    }
}

/**
 * Sync action - executes `mur sync all`
 */
class MurmurSyncAction : MurmurAction(
    "Sync",
    listOf("mur", "sync", "all")
)

/**
 * Extract action - executes `mur learn extract --auto`
 */
class MurmurExtractAction : MurmurAction(
    "Extract",
    listOf("mur", "learn", "extract", "--auto")
)

/**
 * Stats action - executes `mur stats`
 */
class MurmurStatsAction : MurmurAction(
    "Stats",
    listOf("mur", "stats")
)

/**
 * Patterns action - executes `mur patterns`
 */
class MurmurPatternsAction : MurmurAction(
    "Patterns",
    listOf("mur", "patterns")
)
