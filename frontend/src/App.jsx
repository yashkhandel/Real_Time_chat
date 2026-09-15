import { useState } from 'react'
import AuthPage from './pages/AuthPage'
import ChatPage from './pages/ChatPage'

export default function App() {
  const [user, setUser] = useState(null)

  const handleLogout = () => {
    localStorage.removeItem('wire_token')
    setUser(null)
  }

  if (!user) {
    return <AuthPage onAuthenticated={setUser} />
  }

  return <ChatPage user={user} onLogout={handleLogout} />
}
