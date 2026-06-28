<script lang="ts">
  import { api } from '../api'
  import { token } from '../auth'

  let email = ''
  let password = ''
  let error = ''
  let loading = false

  async function handleLogin() {
    if (!email || !password) return
    loading = true
    error = ''
    try {
      const res = await api.auth.login(email, password)
      token.set(res.token)
      window.location.hash = '/projects'
    } catch (e: any) {
      error = e.message || 'Invalid credentials'
    } finally {
      loading = false
    }
  }
</script>

<div class="login-wrap">
  <div class="login-card">
    <div class="login-logo">
      <span>OpenWard</span>
    </div>

    <h1 class="login-title">Sign in</h1>
    <p class="login-sub">Access your OpenWard admin dashboard</p>

    {#if error}
      <div class="alert alert-danger">{error}</div>
    {/if}

    <div class="form-group">
      <label class="form-label" for="email">Email</label>
      <input
        id="email"
        type="email"
        placeholder="admin@example.com"
        bind:value={email}
        on:keydown={e => e.key === 'Enter' && handleLogin()}
      />
    </div>

    <div class="form-group">
      <label class="form-label" for="password">Password</label>
      <input
        id="password"
        type="password"
        placeholder="••••••••"
        bind:value={password}
        on:keydown={e => e.key === 'Enter' && handleLogin()}
      />
    </div>

    <button class="btn btn-primary" style="width:100%; justify-content:center" on:click={handleLogin} disabled={loading}>
      {loading ? 'Signing in…' : 'Sign in'}
    </button>
  </div>
</div>

<style>
  .login-wrap {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg);
    padding: 20px;
  }

  .login-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    padding: 36px;
    width: 100%;
    max-width: 380px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .login-logo {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 4px;
  }

  .login-title {
    font-size: 20px;
    font-weight: 600;
    letter-spacing: -0.02em;
    margin-bottom: -8px;
  }

  .login-sub {
    font-size: 13px;
    color: var(--text-2);
    margin-bottom: 4px;
  }
</style>