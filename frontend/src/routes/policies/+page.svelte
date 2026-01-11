<script lang="ts">
	// Policy management page
</script>

<div class="policies">
	<div class="page-header">
		<h1>Policies</h1>
		<button class="btn-primary">Create Policy</button>
	</div>

	<div class="policy-list">
		<div class="policy-card">
			<div class="policy-header">
				<h3>Data Access Control</h3>
				<span class="badge active">Active</span>
			</div>
			<p class="policy-description">Controls read/write access to sensitive datasets</p>
			<div class="policy-meta">
				<span class="type">Type: Access Control</span>
				<span>Applied to: 12 resources</span>
			</div>
		</div>

		<div class="policy-card">
			<div class="policy-header">
				<h3>Data Quality Rules</h3>
				<span class="badge active">Active</span>
			</div>
			<p class="policy-description">Validates data completeness and accuracy thresholds</p>
			<div class="policy-meta">
				<span class="type">Type: Quality</span>
				<span>Applied to: 8 resources</span>
			</div>
		</div>

		<div class="policy-card">
			<div class="policy-header">
				<h3>PII Protection</h3>
				<span class="badge active">Active</span>
			</div>
			<p class="policy-description">Masks and encrypts personally identifiable information</p>
			<div class="policy-meta">
				<span class="type">Type: Privacy</span>
				<span>Applied to: 5 resources</span>
			</div>
		</div>

		<div class="policy-card">
			<div class="policy-header">
				<h3>GDPR Compliance</h3>
				<span class="badge active">Active</span>
			</div>
			<p class="policy-description">Enforces GDPR data handling requirements</p>
			<div class="policy-meta">
				<span class="type">Type: Compliance</span>
				<span>Applied to: 15 resources</span>
			</div>
		</div>
	</div>
</div>

<style>
	.policies {
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

	.policy-list {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 1.5rem;
	}

	.policy-card {
		background: white;
		padding: 1.5rem;
		border-radius: 8px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		transition: transform 0.2s, box-shadow 0.2s;
	}

	.policy-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		cursor: pointer;
	}

	.policy-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.policy-header h3 {
		margin: 0;
		color: #1a1a1a;
	}

	.policy-description {
		margin: 0 0 1rem 0;
		color: #666;
		line-height: 1.5;
	}

	.policy-meta {
		display: flex;
		gap: 1rem;
		font-size: 0.85rem;
		color: #888;
	}

	.type {
		color: #007bff;
		font-weight: 600;
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
