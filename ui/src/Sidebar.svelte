<script lang="ts">
  import { LayoutGrid, LogOut, ExternalLink } from '@lucide/svelte'
  import { token } from './auth'

  export let currentPage: string = ''

  function logout() {
    token.clear()
    window.location.hash = '/login'
  }
</script>

<aside class="sidebar">
  <div class="sidebar-logo">
    OpenWard
  </div>

  <nav class="sidebar-nav">
    <div class="sidebar-section">Main</div>
    <a href="#/projects" class:active={currentPage === 'projects'}>
      <LayoutGrid size={15} />
      Projects
    </a>
  </nav>

  <div class="sidebar-footer">
    <a href="https://openward.kousic.in" target="_blank" rel="noreferrer">
      <ExternalLink size={15} />
      Documentation
    </a>

    <button on:click={logout}>
      <LogOut size={15} />
      Sign out
    </button>
  </div>
</aside>
