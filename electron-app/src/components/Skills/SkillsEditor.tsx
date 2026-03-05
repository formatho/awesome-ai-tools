import { useState, useEffect } from 'react'
import { Save, Trash2, Plus, FileText, AlertCircle, Check, ChevronDown, ChevronUp } from 'lucide-react'

interface Skill {
  id: string
  name: string
  description: string
  permissions: string[]
  content: string
  createdAt: string
  updatedAt: string
}

const defaultSkillTemplate = `# Skill: Example Skill

## Description
Brief description of what this skill does.

## Capabilities
- Capability 1
- Capability 2
- Capability 3

## Permissions Required
- file.read
- file.write
- http.get

## Usage

### Parameters
\`\`\`json
{
  "param1": "value1",
  "param2": "value2"
}
\`\`\`

### Example
\`\`\`typescript
const result = await agent.executeSkill('example-skill', {
  param1: 'value1',
  param2: 'value2'
})
\`\`\`

## Implementation Notes
- Note 1
- Note 2

## Error Handling
Describe how errors are handled.

## Testing
Describe how to test this skill.

## Security Considerations
- Security note 1
- Security note 2

## Changelog
- **v1.0.0** - Initial implementation
`

const PERMISSIONS = [
  { value: 'file.read', description: 'Read files' },
  { value: 'file.write', description: 'Write files' },
  { value: 'file.delete', description: 'Delete files' },
  { value: 'http.get', description: 'Make GET requests' },
  { value: 'http.post', description: 'Make POST requests' },
  { value: 'shell.run', description: 'Execute shell commands' },
  { value: 'web.fetch', description: 'Fetch web pages' },
  { value: 'web.search', description: 'Search the web' },
  { value: 'notify.desktop', description: 'Send desktop notifications' },
  { value: 'notify.email', description: 'Send emails' },
  { value: 'git.clone', description: 'Clone git repositories' },
  { value: 'git.push', description: 'Push to git' },
  { value: 'git.pull', description: 'Pull from git' },
  { value: 'database.read', description: 'Read from database' },
  { value: 'database.write', description: 'Write to database' },
  { value: 'image.generate', description: 'Generate images' },
]

export default function SkillsEditor() {
  const [skills, setSkills] = useState<Skill[]>([])
  const [selectedSkill, setSelectedSkill] = useState<Skill | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [showPermissionDropdown, setShowPermissionDropdown] = useState(false)
  const [saved, setSaved] = useState(false)
  const [error] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState('')

  useEffect(() => {
    // Load skills from localStorage (or backend in production)
    const savedSkills = localStorage.getItem('agent-skills')
    if (savedSkills) {
      setSkills(JSON.parse(savedSkills))
    } else {
      // Create a sample skill
      const sampleSkill: Skill = {
        id: 'sample-skill',
        name: 'Example Skill',
        description: 'A sample skill to demonstrate the editor',
        permissions: ['file.read', 'file.write'],
        content: defaultSkillTemplate,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      }
      setSkills([sampleSkill])
      localStorage.setItem('agent-skills', JSON.stringify([sampleSkill]))
    }
  }, [])

  const saveSkills = (updatedSkills: Skill[]) => {
    localStorage.setItem('agent-skills', JSON.stringify(updatedSkills))
    setSkills(updatedSkills)
  }

  const createNewSkill = () => {
    const newSkill: Skill = {
      id: `skill-${Date.now()}`,
      name: 'New Skill',
      description: 'Description of the new skill',
      permissions: [],
      content: defaultSkillTemplate,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    }
    const updatedSkills = [...skills, newSkill]
    saveSkills(updatedSkills)
    setSelectedSkill(newSkill)
    setIsEditing(true)
  }

  const updateSkill = (updatedSkill: Skill) => {
    updatedSkill.updatedAt = new Date().toISOString()
    const updatedSkills = skills.map(s => s.id === updatedSkill.id ? updatedSkill : s)
    saveSkills(updatedSkills)
    setSelectedSkill(updatedSkill)
  }

  const deleteSkill = (skillId: string) => {
    if (confirm('Are you sure you want to delete this skill?')) {
      const updatedSkills = skills.filter(s => s.id !== skillId)
      saveSkills(updatedSkills)
      if (selectedSkill?.id === skillId) {
        setSelectedSkill(null)
        setIsEditing(false)
      }
    }
  }

  const togglePermission = (permission: string) => {
    if (!selectedSkill) return
    const permissions = selectedSkill.permissions.includes(permission)
      ? selectedSkill.permissions.filter(p => p !== permission)
      : [...selectedSkill.permissions, permission]
    updateSkill({ ...selectedSkill, permissions })
  }

  const handleSave = () => {
    if (!selectedSkill) return
    updateSkill(selectedSkill)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const filteredSkills = skills.filter(skill =>
    skill.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    skill.description.toLowerCase().includes(searchTerm.toLowerCase())
  )

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-text-primary">Skills Editor</h2>
          <p className="text-text-secondary mt-1">Create and manage agent skills in markdown</p>
        </div>
        <div className="flex gap-2">
          <button onClick={createNewSkill} className="btn-primary">
            <Plus className="w-4 h-4 mr-2" />
            New Skill
          </button>
        </div>
      </div>

      {/* Status Messages */}
      {saved && (
        <div className="bg-green-500/10 border border-green-500/20 rounded-lg p-3 flex items-center gap-2">
          <Check className="w-4 h-4 text-green-500" />
          <span className="text-sm text-green-500">Skill saved successfully!</span>
        </div>
      )}

      {error && (
        <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-3 flex items-center gap-2">
          <AlertCircle className="w-4 h-4 text-red-500" />
          <span className="text-sm text-red-500">{error}</span>
        </div>
      )}

      <div className="grid grid-cols-3 gap-6">
        {/* Skills List */}
        <div className="col-span-1 space-y-4">
          {/* Search */}
          <input
            type="text"
            placeholder="Search skills..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="input w-full"
          />

          {/* Skills List */}
          <div className="space-y-2">
            {filteredSkills.map(skill => (
              <div
                key={skill.id}
                onClick={() => {
                  setSelectedSkill(skill)
                  setIsEditing(false)
                }}
                className={`p-3 rounded-lg border cursor-pointer transition-colors ${
                  selectedSkill?.id === skill.id
                    ? 'border-accent bg-accent/10'
                    : 'border-border hover:border-border-light'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-2">
                    <FileText className="w-4 h-4 mt-0.5 text-text-muted" />
                    <div>
                      <div className="font-medium text-text-primary">{skill.name}</div>
                      <div className="text-xs text-text-secondary mt-1">{skill.description}</div>
                      <div className="flex flex-wrap gap-1 mt-2">
                        {skill.permissions.slice(0, 3).map(perm => (
                          <span
                            key={perm}
                            className="text-xs px-2 py-0.5 bg-surface-hover rounded text-text-muted"
                          >
                            {perm}
                          </span>
                        ))}
                        {skill.permissions.length > 3 && (
                          <span className="text-xs px-2 py-0.5 bg-surface-hover rounded text-text-muted">
                            +{skill.permissions.length - 3}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      deleteSkill(skill.id)
                    }}
                    className="text-text-muted hover:text-red-500 transition-colors"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}

            {filteredSkills.length === 0 && (
              <div className="text-center py-8 text-text-muted">
                <FileText className="w-12 h-12 mx-auto mb-2 opacity-50" />
                <p>No skills found</p>
              </div>
            )}
          </div>
        </div>

        {/* Editor */}
        <div className="col-span-2">
          {selectedSkill ? (
            <div className="space-y-4">
              {/* Edit Mode Toggle */}
              <div className="flex items-center justify-between">
                <div className="flex gap-2">
                  <button
                    onClick={() => setIsEditing(!isEditing)}
                    className={`btn ${isEditing ? 'btn-primary' : 'btn-secondary'}`}
                  >
                    {isEditing ? 'Editing' : 'View'}
                  </button>
                </div>
                {isEditing && (
                  <button onClick={handleSave} className="btn-primary">
                    <Save className="w-4 h-4 mr-2" />
                    Save
                  </button>
                )}
              </div>

              {/* Skill Metadata */}
              {isEditing && (
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm text-text-secondary mb-1">Skill Name</label>
                    <input
                      type="text"
                      value={selectedSkill.name}
                      onChange={(e) => setSelectedSkill({ ...selectedSkill, name: e.target.value })}
                      className="input w-full"
                    />
                  </div>
                  <div>
                    <label className="block text-sm text-text-secondary mb-1">Description</label>
                    <input
                      type="text"
                      value={selectedSkill.description}
                      onChange={(e) => setSelectedSkill({ ...selectedSkill, description: e.target.value })}
                      className="input w-full"
                    />
                  </div>
                </div>
              )}

              {/* Permissions */}
              {isEditing && (
                <div>
                  <label className="block text-sm text-text-secondary mb-2">Permissions</label>
                  <div className="relative">
                    <button
                      onClick={() => setShowPermissionDropdown(!showPermissionDropdown)}
                      className="input w-full text-left flex items-center justify-between"
                    >
                      <span>{selectedSkill.permissions.length} permissions selected</span>
                      {showPermissionDropdown ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                    </button>

                    {showPermissionDropdown && (
                      <div className="absolute z-10 w-full mt-1 bg-surface border border-border rounded-lg shadow-lg max-h-60 overflow-y-auto">
                        {PERMISSIONS.map(perm => (
                          <label
                            key={perm.value}
                            className="flex items-center gap-2 p-2 hover:bg-surface-hover cursor-pointer"
                          >
                            <input
                              type="checkbox"
                              checked={selectedSkill.permissions.includes(perm.value)}
                              onChange={() => togglePermission(perm.value)}
                              className="rounded border-border"
                            />
                            <div>
                              <div className="text-sm text-text-primary font-mono">{perm.value}</div>
                              <div className="text-xs text-text-muted">{perm.description}</div>
                            </div>
                          </label>
                        ))}
                      </div>
                    )}
                  </div>

                  {/* Selected Permissions */}
                  <div className="flex flex-wrap gap-2 mt-2">
                    {selectedSkill.permissions.map(perm => (
                      <span
                        key={perm}
                        className="text-xs px-2 py-1 bg-accent/20 text-accent rounded flex items-center gap-1"
                      >
                        {perm}
                        <button
                          onClick={() => togglePermission(perm)}
                          className="hover:text-red-500"
                        >
                          ×
                        </button>
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Markdown Editor */}
              <div>
                <label className="block text-sm text-text-secondary mb-2">
                  Skill Definition (Markdown)
                </label>
                {isEditing ? (
                  <textarea
                    value={selectedSkill.content}
                    onChange={(e) => setSelectedSkill({ ...selectedSkill, content: e.target.value })}
                    className="input w-full h-96 font-mono text-sm"
                    placeholder="Write your skill definition in markdown..."
                  />
                ) : (
                  <div className="bg-surface-hover border border-border rounded-lg p-4 h-96 overflow-y-auto">
                    <div className="prose prose-invert max-w-none">
                      <pre className="whitespace-pre-wrap text-sm text-text-primary">
                        {selectedSkill.content}
                      </pre>
                    </div>
                  </div>
                )}
              </div>

              {/* Metadata */}
              <div className="text-xs text-text-muted flex justify-between">
                <span>Created: {new Date(selectedSkill.createdAt).toLocaleString()}</span>
                <span>Updated: {new Date(selectedSkill.updatedAt).toLocaleString()}</span>
              </div>
            </div>
          ) : (
            <div className="text-center py-16 text-text-muted">
              <FileText className="w-16 h-16 mx-auto mb-4 opacity-50" />
              <p className="text-lg">Select a skill to view or edit</p>
              <p className="text-sm mt-2">Or create a new skill to get started</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
