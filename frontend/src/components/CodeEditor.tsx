import React from 'react'
import Editor from '@monaco-editor/react'

interface CodeEditorProps {
  value: string
  onChange?: (value: string) => void
  language?: string
  readOnly?: boolean
  height?: number | string
}

export const CodeEditor: React.FC<CodeEditorProps> = ({
  value,
  onChange,
  language = 'json',
  readOnly = false,
  height = 200,
}) => {
  const handleChange = (newValue: string | undefined) => {
    onChange?.(newValue || '')
  }

  return (
    <Editor
      height={height}
      language={language}
      value={value}
      onChange={handleChange}
      options={{
        minimap: { enabled: false },
        fontSize: 13,
        lineNumbers: 'on',
        scrollBeyondLastLine: false,
        automaticLayout: true,
        tabSize: 2,
        wordWrap: 'on',
        readOnly,
        folding: true,
        formatOnPaste: true,
        formatOnType: true,
      }}
      theme="vs-light"
    />
  )
}
