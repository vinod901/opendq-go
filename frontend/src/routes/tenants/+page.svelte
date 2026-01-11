<script lang="ts">
	// Tenant management page
</script>

<div class="tenants">
	<div class="page-header">
		<h1>Tenants</h1>
		<button class="btn-primary">Create Tenant</button>
	</div>

	<div class="tenant-list">
		<div class="tenant-card">
			<h3>Acme Corporation</h3>
			<p class="tenant-slug">acme-corp</p>
			<div class="tenant-meta">
				<span class="badge active">Active</span>
				<span>12 Users</span>
				<span>5 Policies</span>
			</div>
		</div>

		<div class="tenant-card">
			<h3>TechStart Inc.</h3>
			<p class="tenant-slug">techstart</p>
			<div class="tenant-meta">
				<span class="badge active">Active</span>
				<span>8 Users</span>
				<span>3 Policies</span>
			</div>
		</div>

		<div class="tenant-card">
			<h3>Global Industries</h3>
			<p class="tenant-slug">global-ind</p>
			<div class="tenant-meta">
				<span class="badge active">Active</span>
				<span>25 Users</span>
				<span>12 Policies</span>
			</div>
		</div>
	</div>
</div>

<style>
	.tenants {
		padding: 1rem;
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
	}

	h1 {
		margin: 0;
		color: #333;
	}

	.btn-primary {
		background-color: #007bff;
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 4px;
		font-size: 1rem;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.btn-primary:hover {
		background-color: #0056b3;
	}

	.tenant-list {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1.5rem;
	}

	.tenant-card {
		background: white;
		padding: 1.5rem;
		border-radius: 8px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		transition: transform 0.2s, box-shadow 0.2s;
	}

	.tenant-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		cursor: pointer;
	}

	.tenant-card h3 {
		margin: 0 0 0.5rem 0;
		color: #1a1a1a;
	}

	.tenant-slug {
		margin: 0 0 1rem 0;
		color: #666;
		font-size: 0.9rem;
		font-family: monospace;
	}

	.tenant-meta {
		display: flex;
		gap: 1rem;
		font-size: 0.85rem;
		color: #666;
	}

	.badge {
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.badge.active {
		background-color: #28a745;
		color: white;
	}
</style>
