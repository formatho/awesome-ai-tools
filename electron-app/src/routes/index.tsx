import { Routes, Route } from 'react-router-dom'
import Dashboard from '../components/Dashboard/Dashboard'
import AgentList from '../components/Agents/AgentList'
import AgentDetail from '../components/Agents/AgentDetail'
import ChatPage from '../components/Chat/ChatPage'
import TODOList from '../components/TODOs/TODOList'
import CronList from '../components/Cron/CronList'
import ConfigEditor from '../components/Config/ConfigEditor'
import SkillsEditor from '../components/Skills/SkillsEditor'

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Dashboard />} />
      <Route path="/agents" element={<AgentList />} />
      <Route path="/agents/:id/chat" element={<ChatPage />} />
      <Route path="/agents/:id" element={<AgentDetail />} />
      <Route path="/todos" element={<TODOList />} />
      <Route path="/cron" element={<CronList />} />
      <Route path="/config" element={<ConfigEditor />} />
      <Route path="/skills" element={<SkillsEditor />} />
    </Routes>
  )
}
