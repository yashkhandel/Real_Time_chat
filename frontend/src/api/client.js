import axios from 'axios'

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081/api/v1'

const client = axios.create({ baseURL: BASE_URL })

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('wire_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

export const login = (email, password) =>
  client.post('/auth/login', { email, password }).then((r) => r.data)

export const register = (name, email, password) =>
  client.post('/auth/register', { name, email, password }).then((r) => r.data)

export const listUsers = () => client.get('/users').then((r) => r.data)

export const deleteUser = (userId) =>
  client.delete(`/users/${userId}`).then((r) => r.data)

export const listRooms = () => client.get('/rooms').then((r) => r.data)

export const createRoom = (name, isGroup, memberIds) =>
  client
    .post('/rooms', { name, is_group: isGroup, member_ids: memberIds })
    .then((r) => r.data)

export const deleteRoom = (roomId) =>
  client.delete(`/rooms/${roomId}`).then((r) => r.data)

export const getMessages = (roomId, limit = 50) =>
  client.get(`/rooms/${roomId}/messages`, { params: { limit } }).then((r) => r.data)

export const wsUrl = () => {
  const base = (import.meta.env.VITE_WS_URL || 'ws://localhost:8081/api/v1/ws')
  const token = localStorage.getItem('wire_token')
  return `${base}?token=${token}`
}

export default client
