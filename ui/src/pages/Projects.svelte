<script lang="ts">
  import { onMount } from 'svelte'
  import { Plus, LayoutGrid, X, ExternalLink, Copy, Check } from '@lucide/svelte'
  import { api, type Project, type CreateProjectInput } from '../api'
  import Sidebar from '../Sidebar.svelte'

  let projects: Project[] = []
  let loading = true
  let error = ''
  let showModal = false
  let creating = false
  let createError = ''
  let copiedId = ''

  // form state — windows stored in seconds for UX, converted to ns on submit
  let form = emptyForm()

  function emptyForm() {
    return {
      name: '',
      upstream: '',
      algorithm: 'fixed_window',
      enabled: true,
      fw_limit: null as number | null,
      fw_window_s: null as number | null,
      sw_limit: null as number | null,
      sw_window_s: null as number | null,
      tb_capacity: null as number | null,
      tb_refill_rate: null as number | null,
    }
  }

  onMount(async () => {
    await load()
  })

  async function load() {
    loading = true
    error = ''
    try {
      projects = await api.projects.list()
    } catch (e: any) {
      error = e.message
    } finally {
      loading = false
    }
  }

  async function createProject() {
    creating = true
    createError = ''
    try {
      const payload: CreateProjectInput = {
        name: form.name,
        upstream: form.upstream,
        algorithm: form.algorithm,
        enabled: form.enabled,
      }
      if (form.algorithm === 'fixed_window') {
        payload.fw_limit = form.fw_limit ?? undefined
        payload.fw_window = form.fw_window_s != null ? form.fw_window_s * 1_000_000_000 : undefined
      } else if (form.algorithm === 'sliding_window') {
        payload.sw_limit = form.sw_limit ?? undefined
        payload.sw_window = form.sw_window_s != null ? form.sw_window_s * 1_000_000_000 : undefined
      } else if (form.algorithm === 'token_bucket') {
        payload.tb_capacity = form.tb_capacity ?? undefined
        payload.tb_refill_rate = form.tb_refill_rate ?? undefined
      }

      const p = await api.projects.create(payload)
      projects = [p, ...projects]
      showModal = false
      form = emptyForm()
    } catch (e: any) {
      createError = e.message
    } finally {
      creating = false
    }
  }

  function openModal() {
    form = emptyForm()
    createError = ''
    showModal = true
  }

  async function copyKey(key: string) {
    await navigator.clipboard.writeText(key)
    copiedId = key
    setTimeout(() => copiedId = '', 2000)
  }

  function algoLabel(a: string) {
    return { fixed_window: 'Fixed Window', sliding_window: 'Sliding Window', token_bucket: 'Token Bucket' }[a] ?? a
  }

  function navigate(id: string) {
    window.location.hash = `/projects/${id}`
  }
</script>

<div class="layout">
  <Sidebar currentPage="projects" />

  <main class="main-content">
    <div class="page-header">
      <div>
        <h1 class="page-title">Projects</h1>
        <p class="page-subtitle">{projects.length} project{projects.length !== 1 ? 's' : ''}</p>
      </div>
      <button class="btn btn-primary" on:click={openModal}>
        <Plus size={15} />
        New project
      </button>
    </div>

    {#if error}
      <div class="alert alert-danger">{error}</div>
    {:else if loading}
      <div class="empty-state">
        <p>Loading…</p>
      </div>
    {:else if projects.length === 0}
      <div class="empty-state">
        <div class="empty-icon">
          <LayoutGrid size={28} />
        </div>
        <h3>No projects yet</h3>
        <p>A project maps an API key to an upstream URL and a rate limiting policy.<br/>Create one to start proxying requests.</p>
        <button class="btn btn-primary" on:click={openModal}>
          <Plus size={15} />
          Create your first project
        </button>
      </div>
    {:else}
      <div class="card">
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Upstream</th>
                <th>Algorithm</th>
                <th>API Key</th>
                <th>Status</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {#each projects as p}
                <tr>
                  <td><span style="font-weight:500">{p.name}</span></td>
                  <td class="text-muted monospace">{p.upstream}</td>
                  <td>
                    <span class="badge badge-neutral">{algoLabel(p.algorithm)}</span>
                  </td>
                  <td>
                    <div class="flex items-center gap-2">
                      <span class="monospace text-muted">{p.api_key.slice(0, 16)}…</span>
                      <button
                        class="btn btn-ghost btn-sm"
                        style="padding:3px 6px"
                        on:click={() => copyKey(p.api_key)}
                        title="Copy API key"
                      >
                        {#if copiedId === p.api_key}
                          <Check size={12} color="var(--success)" />
                        {:else}
                          <Copy size={12} />
                        {/if}
                      </button>
                    </div>
                  </td>
                  <td>
                    {#if p.enabled}
                      <span class="badge badge-success">Active</span>
                    {:else}
                      <span class="badge badge-danger">Disabled</span>
                    {/if}
                  </td>
                  <td>
                    <button class="btn btn-ghost btn-sm" on:click={() => navigate(p.id)}>
                      <ExternalLink size={13} />
                      View
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    {/if}
  </main>
</div>

{#if showModal}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-backdrop" on:click|self={() => showModal = false}>
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">New project</span>
        <button class="btn btn-ghost btn-sm" style="padding:4px" on:click={() => showModal = false}>
          <X size={15} />
        </button>
      </div>

      <div class="modal-body">
        {#if createError}
          <div class="alert alert-danger">{createError}</div>
        {/if}

        <div class="form-group">
          <label class="form-label" for="name">Name</label>
          <input id="name" placeholder="My API" bind:value={form.name} />
        </div>

        <div class="form-group">
          <label class="form-label" for="upstream">Upstream URL</label>
          <input id="upstream" placeholder="https://api.example.com" bind:value={form.upstream} />
          <span class="form-hint">Requests will be forwarded here</span>
        </div>

        <div class="form-group">
          <label class="form-label" for="algo">Algorithm</label>
          <select id="algo" bind:value={form.algorithm}>
            <option value="fixed_window">Fixed Window</option>
            <option value="sliding_window">Sliding Window</option>
            <option value="token_bucket">Token Bucket</option>
          </select>
        </div>

        {#if form.algorithm === 'fixed_window'}
          <div class="flex gap-3">
            <div class="form-group" style="flex:1">
              <label class="form-label" for="fw_limit">Max requests</label>
              <input id="fw_limit" type="number" placeholder="100" bind:value={form.fw_limit} />
            </div>
            <div class="form-group" style="flex:1">
              <label class="form-label" for="fw_window">Window (seconds)</label>
              <input id="fw_window" type="number" placeholder="60" bind:value={form.fw_window_s} />
            </div>
          </div>
        {:else if form.algorithm === 'sliding_window'}
          <div class="flex gap-3">
            <div class="form-group" style="flex:1">
              <label class="form-label" for="sw_limit">Max requests</label>
              <input id="sw_limit" type="number" placeholder="100" bind:value={form.sw_limit} />
            </div>
            <div class="form-group" style="flex:1">
              <label class="form-label" for="sw_window">Window (seconds)</label>
              <input id="sw_window" type="number" placeholder="60" bind:value={form.sw_window_s} />
            </div>
          </div>
        {:else if form.algorithm === 'token_bucket'}
          <div class="flex gap-3">
            <div class="form-group" style="flex:1">
              <label class="form-label" for="tb_capacity">Capacity</label>
              <input id="tb_capacity" type="number" placeholder="100" bind:value={form.tb_capacity} />
            </div>
            <div class="form-group" style="flex:1">
              <label class="form-label" for="tb_rate">Refill rate (req/sec)</label>
              <input id="tb_rate" type="number" step="0.1" placeholder="10" bind:value={form.tb_refill_rate} />
            </div>
          </div>
        {/if}

        <div class="flex items-center gap-2">
          <label class="toggle">
            <input type="checkbox" bind:checked={form.enabled} />
            <span class="toggle-track"></span>
          </label>
          <span class="form-label" style="margin:0">Enable immediately</span>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-ghost" on:click={() => showModal = false}>Cancel</button>
        <button class="btn btn-primary" on:click={createProject} disabled={creating}>
          {creating ? 'Creating…' : 'Create project'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .empty-icon {
    width: 52px;
    height: 52px;
    border-radius: 12px;
    background: var(--accent-bg);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--accent);
    margin-bottom: 4px;
  }

  .empty-state p {
    max-width: 340px;
    text-align: center;
    line-height: 1.6;
  }
</style>