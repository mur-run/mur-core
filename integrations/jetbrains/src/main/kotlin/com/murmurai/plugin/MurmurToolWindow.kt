package com.murmurai.plugin

import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowFactory
import com.intellij.ui.components.JBScrollPane
import com.intellij.ui.content.ContentFactory
import java.awt.BorderLayout
import java.awt.Font
import javax.swing.JPanel
import javax.swing.JTextArea

/**
 * Tool window factory for Murmur AI output.
 * Provides a dedicated panel for viewing command output.
 */
class MurmurToolWindowFactory : ToolWindowFactory {
    
    override fun createToolWindowContent(project: Project, toolWindow: ToolWindow) {
        val murmurToolWindow = MurmurToolWindow(project)
        val content = ContentFactory.getInstance().createContent(
            murmurToolWindow.getContent(),
            "Output",
            false
        )
        toolWindow.contentManager.addContent(content)
    }
}

/**
 * Murmur AI tool window panel.
 */
class MurmurToolWindow(private val project: Project) {
    
    private val outputArea: JTextArea = JTextArea().apply {
        isEditable = false
        font = Font(Font.MONOSPACED, Font.PLAIN, 12)
        text = """
            |Murmur AI - JetBrains Plugin
            |=============================
            |
            |Available commands (Tools > Murmur AI):
            |
            |  • Sync      - Sync all learning data (mur sync all)
            |  • Extract   - Extract patterns from code (mur learn extract --auto)
            |  • Stats     - Show statistics (mur stats)
            |  • Patterns  - List learned patterns (mur patterns)
            |
            |Make sure 'mur' is installed and available in your PATH.
            |
            |Installation:
            |  brew install karajanchang/tap/mur
            |  # or
            |  go install github.com/karajanchang/murmur-ai/cmd/mur@latest
            |
        """.trimMargin()
    }
    
    fun getContent(): JPanel {
        return JPanel(BorderLayout()).apply {
            add(JBScrollPane(outputArea), BorderLayout.CENTER)
        }
    }
    
    /**
     * Append text to the output area.
     */
    fun appendOutput(text: String) {
        outputArea.append("\n$text")
        outputArea.caretPosition = outputArea.document.length
    }
    
    /**
     * Clear the output area.
     */
    fun clearOutput() {
        outputArea.text = ""
    }
}
