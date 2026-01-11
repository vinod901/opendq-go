<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="app">
	<header>
		<nav>
			<div class="nav-brand">
				<h1>OpenDQ Control Plane</h1>
			</div>
			<div class="nav-links">
				<a href="/">Dashboard</a>
				<a href="/tenants">Tenants</a>
				<a href="/policies">Policies</a>
				<a href="/workflows">Workflows</a>
				<a href="/lineage">Lineage</a>
			</div>
		</nav>
	</header>

	<main>
		{@render children()}
	</main>

	<footer>
		<p>OpenDQ Control Plane - Enterprise Data Quality Management</p>
	</footer>
</div>

<style>
	.app {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}

	header {
		background-color: #1a1a1a;
		padding: 1rem 2rem;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	nav {
		display: flex;
		justify-content: space-between;
		align-items: center;
		max-width: 1200px;
		margin: 0 auto;
	}

	.nav-brand h1 {
		margin: 0;
		color: #fff;
		font-size: 1.5rem;
	}

	.nav-links {
		display: flex;
		gap: 2rem;
	}

	.nav-links a {
		color: #ccc;
		text-decoration: none;
		padding: 0.5rem 1rem;
		border-radius: 4px;
		transition: all 0.2s;
	}

	.nav-links a:hover {
		color: #fff;
		background-color: rgba(255, 255, 255, 0.1);
	}

	main {
		flex: 1;
		width: 100%;
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem;
	}

	footer {
		background-color: #1a1a1a;
		color: #ccc;
		padding: 1rem 2rem;
		text-align: center;
	}

	footer p {
		margin: 0;
	}

	:global(body) {
		margin: 0;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
		background-color: #f5f5f5;
	}

	:global(*) {
		box-sizing: border-box;
	}
</style>

