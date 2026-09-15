import { useEffect, useRef, useState, useCallback } from 'react'
import { listRooms, listUsers, createRoom, deleteRoom, deleteUser, getMessages } from '../api/client'
import { useChatSocket } from '../hooks/useChatSocket'

export default function ChatPage({ user, onLogout }) {
  const [rooms, setRooms] = useState([])
  const [users, setUsers] = useState([])
  const [activeRoom, setActiveRoom] = useState(null)
  const [messages, setMessages] = useState([])
  const [draft, setDraft] = useState('')
  const [typingUsers, setTypingUsers] = useState({}) // roomId -> Set of userIds typing
  const [onlineUsers, setOnlineUsers] = useState({}) // userId -> bool
  const [newRoomName, setNewRoomName] = useState('')
  const [selectedMemberIds, setSelectedMemberIds] = useState([])
  const scrollRef = useRef(null)
  const typingTimeoutRef = useRef(null)

  const handleEnvelope = useCallback((env) => {
    if (env.type === 'message') {
      setMessages((prev) =>
        env.room_id === activeRoomRef.current ? [...prev, env.data] : prev
      )
    } else if (env.type === 'typing') {
      const uid = env.data.user_id
      setTypingUsers((prev) => {
        const set = new Set(prev[env.room_id] || [])
        set.add(uid)
        return { ...prev, [env.room_id]: set }
      })
      setTimeout(() => {
        setTypingUsers((prev) => {
          const set = new Set(prev[env.room_id] || [])
          set.delete(uid)
          return { ...prev, [env.room_id]: set }
        })
      }, 3000)
    } else if (env.type === 'presence') {
      setOnlineUsers((prev) => ({ ...prev, [env.data.user_id]: env.data.online }))
    }
  }, [])

  const { connected, send } = useChatSocket(handleEnvelope)

  const activeRoomRef = useRef(null)
  useEffect(() => {
    activeRoomRef.current = activeRoom?.id
  }, [activeRoom])

  useEffect(() => {
    refreshRooms()
    listUsers().then(setUsers)
  }, [])

  const refreshRooms = async () => {
    const data = await listRooms()
    setRooms(data || [])
    if (data?.length && !activeRoom) setActiveRoom(data[0])
  }

  useEffect(() => {
    if (activeRoom) {
      getMessages(activeRoom.id).then(setMessages)
    }
  }, [activeRoom])

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: 'smooth' })
  }, [messages])

  const handleSend = (e) => {
    e.preventDefault()
    if (!draft.trim() || !activeRoom) return
    send({ type: 'message', room_id: activeRoom.id, body: draft.trim() })
    setDraft('')
  }

  const handleTypingInput = (val) => {
    setDraft(val)
    if (!activeRoom) return
    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current)
    send({ type: 'typing', room_id: activeRoom.id })
    typingTimeoutRef.current = setTimeout(() => {}, 2000)
  }

  const handleCreateRoom = async (e) => {
    e.preventDefault()
    if (!newRoomName.trim() || selectedMemberIds.length === 0) return
    const room = await createRoom(newRoomName, selectedMemberIds.length > 1, selectedMemberIds)
    setNewRoomName('')
    setSelectedMemberIds([])
    await refreshRooms()
    setActiveRoom(room)
  }

  const toggleMember = (id) => {
    setSelectedMemberIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    )
  }

  const handleDeleteRoom = async (roomId, e) => {
    e.stopPropagation()
    const confirmed = window.confirm('Delete this room and all its messages? This cannot be undone.')
    if (!confirmed) return
    await deleteRoom(roomId)
    if (activeRoom?.id === roomId) {
      setActiveRoom(null)
      setMessages([])
    }
    await refreshRooms()
  }

  const handleDeleteUser = async (userId, e) => {
    e.stopPropagation()
    const confirmed = window.confirm(
      'Delete this account? This removes them from all rooms and cannot be undone.'
    )
    if (!confirmed) return
    await deleteUser(userId)
    setUsers((prev) => prev.filter((u) => u.id !== userId))
    setSelectedMemberIds((prev) => prev.filter((id) => id !== userId))
  }

  const typingNames = activeRoom
    ? [...(typingUsers[activeRoom.id] || [])]
        .filter((uid) => uid !== user.id)
        .map((uid) => users.find((u) => u.id === uid)?.name)
        .filter(Boolean)
    : []

  return (
    <div style={styles.wrap}>
      <aside style={styles.sidebar}>
        <div style={styles.brandRow}>
          <span style={styles.brand}>WIRE</span>
          <span style={{ ...styles.statusDot, background: connected ? 'var(--accent)' : 'var(--danger)' }} />
        </div>
        <div style={styles.userRow}>
          <span style={{ color: 'var(--text-muted)', fontSize: 13 }}>{user?.name}</span>
          <button style={styles.logoutBtn} onClick={onLogout}>Sign out</button>
        </div>

        <div style={styles.sectionLabel}>Rooms</div>
        <div style={styles.roomList}>
          {rooms.map((r) => (
            <div key={r.id} style={styles.roomRow}>
              <button
                onClick={() => setActiveRoom(r)}
                style={{ ...styles.roomItem, ...(activeRoom?.id === r.id ? styles.roomItemActive : {}) }}
              >
                {r.name}
              </button>
              <button
                style={styles.deleteIconBtn}
                onClick={(e) => handleDeleteRoom(r.id, e)}
                title="Delete room"
              >
                ×
              </button>
            </div>
          ))}
        </div>

        <form onSubmit={handleCreateRoom} style={styles.newRoomForm}>
          <input
            style={styles.smallInput}
            placeholder="New room name"
            value={newRoomName}
            onChange={(e) => setNewRoomName(e.target.value)}
          />
          <div style={styles.memberPicker}>
            {users.filter((u) => u.id !== user.id).map((u) => (
              <div key={u.id} style={styles.memberRow}>
                <label style={styles.memberOption}>
                  <input
                    type="checkbox"
                    checked={selectedMemberIds.includes(u.id)}
                    onChange={() => toggleMember(u.id)}
                  />
                  <span style={{ marginLeft: 6 }}>
                    {u.name} <span style={styles.memberEmail}>({u.email})</span>
                  </span>
                </label>
                <button
                  style={styles.deleteIconBtn}
                  onClick={(e) => handleDeleteUser(u.id, e)}
                  title="Delete this account"
                >
                  ×
                </button>
              </div>
            ))}
          </div>
          <button style={styles.addBtn}>Create room</button>
        </form>
      </aside>

      <main style={styles.main}>
        {activeRoom ? (
          <>
            <div style={styles.header}>
              <div style={styles.pageTitle}>{activeRoom.name}</div>
              <div style={styles.typingIndicator}>
                {typingNames.length > 0 && `${typingNames.join(', ')} typing…`}
              </div>
            </div>

            <div style={styles.messageList} ref={scrollRef}>
              {messages.map((m) => (
                <MessageBubble key={m.id} message={m} isOwn={m.sender_id === user.id} online={onlineUsers[m.sender_id]} />
              ))}
              {messages.length === 0 && (
                <div style={styles.emptyState}>No messages yet. Say hello 👋</div>
              )}
            </div>

            <form onSubmit={handleSend} style={styles.composer}>
              <input
                style={styles.composerInput}
                value={draft}
                onChange={(e) => handleTypingInput(e.target.value)}
                placeholder="Type a message…"
              />
              <button style={styles.sendBtn}>Send</button>
            </form>
          </>
        ) : (
          <div style={styles.emptyState}>Create a room on the left to start chatting.</div>
        )}
      </main>
    </div>
  )
}

function MessageBubble({ message, isOwn, online }) {
  return (
    <div style={{ ...styles.bubbleRow, justifyContent: isOwn ? 'flex-end' : 'flex-start' }}>
      {!isOwn && (
        <div style={{ ...styles.avatar, background: message.sender?.avatar_color || 'var(--accent)' }}>
          {message.sender?.name?.[0]?.toUpperCase()}
          {online && <span style={styles.onlineDot} />}
        </div>
      )}
      <div style={{ ...styles.bubble, ...(isOwn ? styles.bubbleOwn : styles.bubbleOther) }}>
        {!isOwn && <div style={styles.senderName}>{message.sender?.name}</div>}
        <div>{message.body}</div>
        <div style={styles.timestamp}>
          {new Date(message.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
        </div>
      </div>
    </div>
  )
}

const styles = {
  wrap: { display: 'flex', height: '100vh' },
  sidebar: { width: 260, background: 'var(--surface)', borderRight: '1px solid var(--border)', padding: '20px 16px', display: 'flex', flexDirection: 'column' },
  brandRow: { display: 'flex', alignItems: 'center', gap: 8, marginBottom: 16 },
  brand: { fontFamily: 'var(--font-mono)', fontSize: 14, letterSpacing: 3, color: 'var(--accent)' },
  statusDot: { width: 8, height: 8, borderRadius: '50%' },
  userRow: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 },
  logoutBtn: { background: 'none', border: 'none', color: 'var(--text-muted)', fontSize: 12, cursor: 'pointer', textDecoration: 'underline' },
  sectionLabel: { fontSize: 11, letterSpacing: 1, color: 'var(--text-muted)', textTransform: 'uppercase', marginBottom: 8 },
  roomList: { display: 'flex', flexDirection: 'column', gap: 4, maxHeight: 200, overflowY: 'auto' },
  roomRow: { display: 'flex', alignItems: 'center', gap: 2 },
  roomItem: { flex: 1, textAlign: 'left', background: 'none', border: 'none', color: 'var(--text-primary)', padding: '8px 10px', borderRadius: 6, cursor: 'pointer', fontSize: 13 },
  roomItemActive: { background: 'var(--surface-raised)', color: 'var(--accent)' },
  deleteIconBtn: { background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontSize: 16, padding: '4px 8px', lineHeight: 1 },
  newRoomForm: { display: 'flex', flexDirection: 'column', gap: 8, marginTop: 16, borderTop: '1px solid var(--border)', paddingTop: 16 },
  smallInput: { background: 'var(--surface-raised)', border: '1px solid var(--border)', borderRadius: 6, padding: '8px 10px', color: 'var(--text-primary)', fontSize: 13, outline: 'none' },
  memberPicker: { display: 'flex', flexDirection: 'column', gap: 2, maxHeight: 140, overflowY: 'auto', fontSize: 12, color: 'var(--text-muted)' },
  memberRow: { display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 4 },
  memberOption: { display: 'flex', alignItems: 'center', flex: 1, minWidth: 0 },
  memberEmail: { color: 'var(--text-muted)', fontSize: 10.5 },
  addBtn: { background: 'var(--accent)', color: '#04211F', border: 'none', borderRadius: 6, padding: '8px', cursor: 'pointer', fontWeight: 600, fontSize: 12 },
  main: { flex: 1, display: 'flex', flexDirection: 'column' },
  header: { padding: '18px 24px', borderBottom: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' },
  pageTitle: { fontSize: 18, fontWeight: 600 },
  typingIndicator: { fontSize: 12, color: 'var(--accent)', fontStyle: 'italic' },
  messageList: { flex: 1, overflowY: 'auto', padding: '20px 24px', display: 'flex', flexDirection: 'column', gap: 10 },
  bubbleRow: { display: 'flex', gap: 8, alignItems: 'flex-end' },
  avatar: { width: 28, height: 28, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, fontWeight: 700, color: '#04211F', position: 'relative', flexShrink: 0 },
  onlineDot: { position: 'absolute', bottom: -1, right: -1, width: 8, height: 8, borderRadius: '50%', background: 'var(--accent)', border: '2px solid var(--surface)' },
  bubble: { maxWidth: '60%', padding: '8px 12px', borderRadius: 12, fontSize: 14 },
  bubbleOwn: { background: 'var(--accent-dim)', color: 'var(--text-primary)', borderBottomRightRadius: 2 },
  bubbleOther: { background: 'var(--surface-raised)', borderBottomLeftRadius: 2 },
  senderName: { fontSize: 11, color: 'var(--accent)', marginBottom: 2, fontWeight: 600 },
  timestamp: { fontSize: 10, color: 'var(--text-muted)', marginTop: 4, textAlign: 'right' },
  emptyState: { color: 'var(--text-muted)', fontSize: 14, textAlign: 'center', margin: 'auto' },
  composer: { display: 'flex', gap: 10, padding: '16px 24px', borderTop: '1px solid var(--border)' },
  composerInput: { flex: 1, background: 'var(--surface-raised)', border: '1px solid var(--border)', borderRadius: 8, padding: '10px 14px', color: 'var(--text-primary)', fontSize: 14, outline: 'none' },
  sendBtn: { background: 'var(--accent)', color: '#04211F', border: 'none', borderRadius: 8, padding: '10px 18px', fontWeight: 600, cursor: 'pointer' },
}
