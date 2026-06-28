<script lang="ts">
  import { onMount } from 'svelte'
  import { ArrowLeft, Copy, Check, Trash2, X } from '@lucide/svelte'
  import { api, type Project, type RequestLog } from '../api'
  import Sidebar from '../Sidebar.svelte'

  export let id: string

  let project: Project | null = null
  let logs: RequestLog[] = []
  let loading = true
  let logsLoading = true
  let logsError = ''
  let error = ''
  let copiedKey = false
  let showDeleteModal = false
  let deleting = false
  let logOffset = 0
  const logsPageSize = 25

  // edit state
  let editing = false
  let saving = false
  let editError = ''
  let editForm: any = {}

  onMount(async () => {
    await Promise.all([loadProject(), loadLogs()])
  })

  async function loadProject() {
    loading = true
    try {
      project = await api.projects.get(id)
      resetEditForm()
    } catch (e: any) {
      error = e.message
    } finally {
      loading = false
    }
  }

  async function loadLogs(offset = logOffset) {
    logOffset = offset
    logsLoading = true
    logsError = ''
    try {
      logs = await api.logs.list(id, logsPageSize, offset)
    } catch (e: any) {
      logs = []
      logsError = e.message
    } finally {
      logsLoading = false
    }
  }

  function previousLogs() {
    if (logOffset === 0) return
    loadLogs(Math.max(0, logOffset - logsPageSize))
  }

  function nextLogs() {
    if (!hasMoreLogs) return
    loadLogs(logOffset + logsPageSize)
  }

  function resetEditForm() {
    if (!project) return
    editForm = {
      name: project.name,
      upstream: project.upstream,
      enabled: project.enabled,
      // windows converted ns → seconds for display
      fw_limit: project.fw_limit ?? null,
      fw_window_s: project.fw_window != null ? project.fw_window / 1_000_000_000 : null,
      sw_limit: project.sw_limit ?? null,
      sw_window_s: project.sw_window != null ? project.sw_window / 1_000_000_000 : null,
      tb_capacity: project.tb_capacity ?? null,
      tb_refill_rate: project.tb_refill_rate ?? null,
    }
  }

  async function saveProject() {
    if (!project) return
    saving = true
    editError = ''
    try {
      const payload: any = {
        name: editForm.name,
        upstream: editForm.upstream,
        enabled: editForm.enabled,
      }
      if (project.algorithm === 'fixed_window') {
        payload.fw_limit = editForm.fw_limit
        payload.fw_window = editForm.fw_window_s != null ? editForm.fw_window_s * 1_000_000_000 : null
      } else if (project.algorithm === 'sliding_window') {
        payload.sw_limit = editForm.sw_limit
        payload.sw_window = editForm.sw_window_s != null ? editForm.sw_window_s * 1_000_000_000 : null
      } else if (project.algorithm === 'token_bucket') {
        payload.tb_capacity = editForm.tb_capacity
        payload.tb_refill_rate = editForm.tb_refill_rate
      }
      project = await api.projects.update(id, payload)
      editing = false
    } catch (e: any) {
      editError = e.message
    } finally {
      saving = false
    }
  }

  async function deleteProject() {
    deleting = true
    try {
      await api.projects.delete(id)
      window.location.hash = '/projects'
    } catch (e: any) {
      deleting = false
    }
  }

  async function copyKey() {
    if (!project) return
    await navigator.clipboard.writeText(project.api_key)
    copiedKey = true
    setTimeout(() => copiedKey = false, 2000)
  }

  function algoLabel(a: string) {
    return { fixed_window: 'Fixed Window', sliding_window: 'Sliding Window', token_bucket: 'Token Bucket' }[a] ?? a
  }

  function formatDate(s: string) {
    return new Date(s).toLocaleString()
  }

  function nsToDuration(ns: number): string {
    const s = ns / 1_000_000_000
    if (s >= 3600) return `${s / 3600}h`
    if (s >= 60) return `${s / 60}m`
    return `${s}s`
  }

  $: totalRequests = logs.length
  $: allowedRequests = logs.filter(l => l.allowed).length
  $: blockedRequests = logs.filter(l => !l.allowed).length
  $: allowRate = totalRequests > 0 ? Math.round((allowedRequests / totalRequests) * 100) : 0
  $: hasMoreLogs = logs.length === logsPageSize
  $: currentLogPage = Math.floor(logOffset / logsPageSize) + 1
  $: firstLogIndex = logOffset + 1
  $: lastLogIndex = logOffset + logs.length
</script>

<div class="layout">
  <Sidebar currentPage="projects" />

  <main class="main-content">
    {#if loading}
      <div class="empty-state"><p>Loading…</p></div>
    {:else if error}
      <div class="alert alert-danger">{error}</div>
    {:else if project}
      <!-- Header -->
      <div class="page-header">
        <div class="flex items-center gap-3">
          <a href="#/projects" class="btn btn-ghost btn-sm" style="padding:5px 8px">
            <ArrowLeft size={15} />
          </a>
          <div>
            <h1 class="page-title">{project.name}</h1>
            <p class="page-subtitle">{project.upstream}</p>
          </div>
          {#if project.enabled}
            <span class="badge badge-success">Active</span>
          {:else}
            <span class="badge badge-danger">Disabled</span>
          {/if}
        </div>
        <div class="flex items-center gap-2">
          <button class="btn btn-ghost btn-sm" on:click={() => { editing = true; resetEditForm() }}>
            Edit
          </button>
          <button class="btn btn-danger btn-sm" on:click={() => showDeleteModal = true}>
            <Trash2 size={13} />
            Delete
          </button>
        </div>
      </div>

      <!-- Stats -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-label">Requests shown</div>
          <div class="stat-value">{totalRequests}</div>
          <div class="stat-sub">Current page only</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Allowed</div>
          <div class="stat-value" style="color:var(--success)">{allowedRequests}</div>
          <div class="stat-sub">{allowRate}% pass rate</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Blocked</div>
          <div class="stat-value" style="color:var(--danger)">{blockedRequests}</div>
          <div class="stat-sub">Rate limited</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Algorithm</div>
          <div class="stat-value" style="font-size:16px;margin-top:4px">{algoLabel(project.algorithm)}</div>
        </div>
      </div>

      <!-- Rate limit config -->
      <div class="card" style="margin-bottom:20px">
        <div class="card-header">
          <span class="card-title">Rate limit config</span>
          <span class="badge badge-neutral">{algoLabel(project.algorithm)}</span>
        </div>
        <div class="card-body">
          <div class="config-grid">
            {#if project.algorithm === 'fixed_window'}
              <div class="config-item">
                <div class="config-label">Max requests</div>
                <div class="config-value">{project.fw_limit ?? '—'}</div>
              </div>
              <div class="config-item">
                <div class="config-label">Window</div>
                <div class="config-value">{project.fw_window != null ? nsToDuration(project.fw_window) : '—'}</div>
              </div>
            {:else if project.algorithm === 'sliding_window'}
              <div class="config-item">
                <div class="config-label">Max requests</div>
                <div class="config-value">{project.sw_limit ?? '—'}</div>
              </div>
              <div class="config-item">
                <div class="config-label">Window</div>
                <div class="config-value">{project.sw_window != null ? nsToDuration(project.sw_window) : '—'}</div>
              </div>
            {:else if project.algorithm === 'token_bucket'}
              <div class="config-item">
                <div class="config-label">Capacity</div>
                <div class="config-value">{project.tb_capacity ?? '—'}</div>
              </div>
              <div class="config-item">
                <div class="config-label">Refill rate</div>
                <div class="config-value">{project.tb_refill_rate != null ? `${project.tb_refill_rate} req/s` : '—'}</div>
              </div>
            {/if}
          </div>
        </div>
      </div>

      <!-- API Key -->
      <div class="card" style="margin-bottom:20px">
        <div class="card-header">
          <span class="card-title">API Key</span>
        </div>
        <div class="card-body">
          <div class="flex items-center gap-2">
            <input
              class="monospace"
              style="flex:1; background:var(--bg); color:var(--text-2)"
              readonly
              value={project.api_key}
            />
            <button class="btn btn-ghost" on:click={copyKey}>
              {#if copiedKey}
                <Check size={14} color="var(--success)" /> Copied
              {:else}
                <Copy size={14} /> Copy
              {/if}
            </button>
          </div>
          <p class="form-hint" style="margin-top:8px">Pass this as the <span class="monospace">X-API-Key</span> header on requests to the proxy.</p>
        </div>
      </div>

      <!-- Logs -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">Recent requests</span>
          <div class="flex items-center gap-2">
            <span class="text-muted text-sm">
              Page {currentLogPage}
            </span>
            <button class="btn btn-ghost btn-sm" on:click={() => loadLogs()}>
              Refresh
            </button>
          </div>
        </div>
        {#if logsError}
          <div class="alert alert-danger" style="margin:20px">{logsError}</div>
        {/if}
        {#if logsLoading}
          <div class="empty-state" style="padding:40px"><p>Loading…</p></div>
        {:else if logs.length === 0}
          <div class="empty-state">
            <h3>No requests yet</h3>
            <p>Requests will appear here once traffic hits this project.</p>
          </div>
        {:else}
          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Request ID</th>
                  <th>Time</th>
                  <th>Status</th>
                  <th>HTTP Code</th>
                </tr>
              </thead>
              <tbody>
                {#each logs as log}
                  <tr>
                    <td class="monospace text-sm">#{log.id}</td>
                    <td class="text-muted monospace text-sm">{formatDate(log.requested_at)}</td>
                    <td>
                      {#if log.allowed}
                        <span class="badge badge-success">Allowed</span>
                      {:else}
                        <span class="badge badge-danger">Blocked</span>
                      {/if}
                    </td>
                    <td class="monospace text-sm">
                      {#if log.status_code != null}
                        {log.status_code}
                      {:else}
                        <span class="text-muted">—</span>
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
          <div class="pagination">
            <div class="pagination-meta text-muted text-sm">
              Showing {firstLogIndex}-{lastLogIndex}
            </div>
            <div class="pagination-actions">
              <button class="btn btn-ghost btn-sm" on:click={previousLogs} disabled={logOffset === 0 || logsLoading}>
                Previous
              </button>
              <button class="btn btn-ghost btn-sm" on:click={nextLogs} disabled={!hasMoreLogs || logsLoading}>
                Next
              </button>
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </main>
</div>

<!-- Edit modal -->
{#if editing && project}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="modal-backdrop" on:click|self={() => editing = false}>
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">Edit project</span>
        <button class="btn btn-ghost btn-sm" style="padding:4px" on:click={() => editing = false}>
          <X size={15} />
        </button>
      </div>
      <div class="modal-body">
        {#if editError}
          <div class="alert alert-danger">{editError}</div>
        {/if}
        <div class="form-group">
          <label class="form-label" for="edit-name">Name</label>
          <input id="edit-name" bind:value={editForm.name} />
        </div>
        <div class="form-group">
          <label class="form-label" for="edit-upstream">Upstream URL</label>
          <input id="edit-upstream" bind:value={editForm.upstream} />
        </div>

        <div class="divider"></div>

        {#if project.algorithm === 'fixed_window'}
          <div class="flex gap-3">
            <div class="form-group" style="flex:1">
              <label class="form-label" for="edit-fw-limit">Max requests</label>
              <input id="edit-fw-limit" type="number" bind:value={editForm.fw_limit} />
            </div>
            <div class="form-group" style="flex:1">
              <label class="form-label" for="edit-fw-window">Window (seconds)</label>
              <input id="edit-fw-window" type="number" bind:value={editForm.fw_window_s} />
            </div>
          </div>
        {:else if project.algorithm === 'sliding_window'}
          <div class="flex gap-3">
            <div class="form-group" style="flex:1">
              <label class="form-label" for="edit-sw-limit">Max requests</label>
              <input id="edit-sw-limit" type="number" bind:value={editForm.sw_limit} />
            </div>
            <div class="form-group" style="flex:1">
              <label class="form-label" for="edit-sw-window">Window (seconds)</label>
              <input id="edit-sw-window" type="number" bind:value={editForm.sw_window_s} />
            </div>
          </div>
        {:else if project.algorithm === 'token_bucket'}
          <div class="flex gap-3">
            <div class="form-group" style="flex:1">
              <label class="form-label" for="edit-tb-capacity">Capacity</label>
              <input id="edit-tb-capacity" type="number" bind:value={editForm.tb_capacity} />
            </div>
            <div class="form-group" style="flex:1">
              <label class="form-label" for="edit-tb-rate">Refill rate (req/sec)</label>
              <input id="edit-tb-rate" type="number" step="0.1" bind:value={editForm.tb_refill_rate} />
            </div>
          </div>
        {/if}

        <div class="flex items-center gap-2">
          <label class="toggle">
            <input type="checkbox" bind:checked={editForm.enabled} />
            <span class="toggle-track"></span>
          </label>
          <span class="form-label" style="margin:0">Enabled</span>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-ghost" on:click={() => editing = false}>Cancel</button>
        <button class="btn btn-primary" on:click={saveProject} disabled={saving}>
          {saving ? 'Saving…' : 'Save changes'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete confirmation -->
{#if showDeleteModal}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="modal-backdrop" on:click|self={() => showDeleteModal = false}>
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">Delete project</span>
        <button class="btn btn-ghost btn-sm" style="padding:4px" on:click={() => showDeleteModal = false}>
          <X size={15} />
        </button>
      </div>
      <div class="modal-body">
        <p>Are you sure you want to delete <strong>{project?.name}</strong>? This will remove all request logs and cannot be undone.</p>
      </div>
      <div class="modal-footer">
        <button class="btn btn-ghost" on:click={() => showDeleteModal = false}>Cancel</button>
        <button class="btn btn-danger" on:click={deleteProject} disabled={deleting}>
          {deleting ? 'Deleting…' : 'Delete project'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .config-grid {
    display: flex;
    gap: 40px;
  }

  .config-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .config-label {
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-2);
  }

  .config-value {
    font-size: 22px;
    font-weight: 600;
    letter-spacing: -0.02em;
    color: var(--text);
  }

  .pagination {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 16px 16px;
    border-top: 1px solid var(--border);
  }

  .pagination-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }
</style>
