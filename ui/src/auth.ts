import { writable, derived } from 'svelte/store'

const _token = writable<string>(localStorage.getItem('ow_token') ?? '')

export const token = {
  subscribe: _token.subscribe,
  set: (t: string) => {
    localStorage.setItem('ow_token', t)
    _token.set(t)
  },
  clear: () => {
    localStorage.removeItem('ow_token')
    _token.set('')
  },
}

export const isAuthed = derived(_token, t => t.length > 0)