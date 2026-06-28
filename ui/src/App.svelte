<script lang="ts">
  import { onMount } from 'svelte'
  import { isAuthed } from './auth'
  import Login from './pages/Login.svelte'
  import Projects from './pages/Projects.svelte'
  import ProjectDetail from './pages/ProjectDetail.svelte'

  type Route =
    | { name: 'login' }
    | { name: 'projects' }
    | { name: 'project'; id: string }
    | { name: 'notfound' }

  let route: Route = { name: 'projects' }
  let authed = false
  isAuthed.subscribe(v => authed = v)

  function parse(hash: string): Route {
    const path = hash.replace(/^#/, '') || '/'
    if (path === '/' || path === '/projects') return { name: 'projects' }
    const m = path.match(/^\/projects\/([^/]+)$/)
    if (m) return { name: 'project', id: m[1] }
    if (path === '/login') return { name: 'login' }
    return { name: 'notfound' }
  }

  function navigate() {
    const r = parse(window.location.hash)
    if (!authed && r.name !== 'login') {
      route = { name: 'login' }
      return
    }
    if (authed && r.name === 'login') {
      window.location.hash = '/projects'
      return
    }
    route = r
  }

  onMount(() => {
    navigate()
    window.addEventListener('hashchange', navigate)
    return () => window.removeEventListener('hashchange', navigate)
  })
</script>

{#if route.name === 'login'}
  <Login />
{:else if route.name === 'projects'}
  <Projects />
{:else if route.name === 'project'}
  <ProjectDetail id={route.id} />
{:else}
  <div style="padding:40px;color:var(--text-2)">Page not found.</div>
{/if}