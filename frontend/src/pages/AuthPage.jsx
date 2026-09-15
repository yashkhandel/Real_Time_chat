import { useState } from 'react'
import { login, register } from '../api/client'

export default function AuthPage({ onAuthenticated }) {
  const [mode, setMode] = useState('login')
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = mode === 'login' ? await login(email, password) : await register(name, email, password)
      localStorage.setItem('wire_token', data.token)
      onAuthenticated(data.user)
    } catch (err) {
      setError(err.response?.data?.error || 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={styles.wrap}>
      <div style={styles.card}>
        <div style={styles.eyebrow}>WIRE</div>
        <h1 style={styles.title}>{mode === 'login' ? 'Sign in' : 'Create your account'}</h1>
        <p style={styles.sub}>Realtime messaging, built on WebSockets.</p>

        <form onSubmit={submit} style={styles.form}>
          {mode === 'register' && (
            <input style={styles.input} placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
          )}
          <input style={styles.input} placeholder="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          <input style={styles.input} placeholder="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
          {error && <div style={styles.error}>{error}</div>}
          <button style={styles.button} disabled={loading}>
            {loading ? 'Working…' : mode === 'login' ? 'Sign in' : 'Register'}
          </button>
        </form>

        <button style={styles.switch} onClick={() => setMode(mode === 'login' ? 'register' : 'login')}>
          {mode === 'login' ? "Don't have an account? Register" : 'Already have an account? Sign in'}
        </button>
      </div>
    </div>
  )
}

const styles = {
  wrap: { height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'radial-gradient(circle at 30% 20%, #12233A 0%, #0B0E14 60%)' },
  card: { width: 360, background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 12, padding: '32px 28px' },
  eyebrow: { fontFamily: 'var(--font-mono)', fontSize: 12, letterSpacing: 3, color: 'var(--accent)', marginBottom: 8 },
  title: { fontSize: 22, margin: '0 0 4px 0', fontWeight: 600 },
  sub: { color: 'var(--text-muted)', fontSize: 14, margin: '0 0 24px 0' },
  form: { display: 'flex', flexDirection: 'column', gap: 12 },
  input: { background: 'var(--surface-raised)', border: '1px solid var(--border)', borderRadius: 8, padding: '10px 12px', color: 'var(--text-primary)', fontSize: 14, outline: 'none' },
  error: { color: 'var(--danger)', fontSize: 13 },
  button: { background: 'var(--accent)', color: '#04211F', border: 'none', borderRadius: 8, padding: '10px 12px', fontWeight: 600, fontSize: 14, cursor: 'pointer', marginTop: 4 },
  switch: { background: 'none', border: 'none', color: 'var(--text-muted)', fontSize: 13, marginTop: 18, cursor: 'pointer', textDecoration: 'underline' },
}
