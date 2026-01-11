<script lang="ts">
	// Workflow management page
</script>

<div class="workflows">
	<div class="page-header">
		<h1>Workflows</h1>
		<button class="btn-primary">Create Workflow</button>
	</div>

	<div class="workflow-list">
		<div class="workflow-card">
			<div class="workflow-header">
				<h3>Data Quality Check</h3>
				<span class="status running">Running</span>
			</div>
			<p class="workflow-type">Type: Data Quality</p>
			<div class="workflow-state">
				<div class="state-indicator">
					<span class="state">pending</span>
					<span class="arrow">→</span>
					<span class="state current">validating</span>
					<span class="arrow">→</span>
					<span class="state">completed</span>
				</div>
			</div>
			<div class="workflow-meta">
				<span>Started: 2 hours ago</span>
			</div>
		</div>

		<div class="workflow-card">
			<div class="workflow-header">
				<h3>Approval Process</h3>
				<span class="status pending">Pending</span>
			</div>
			<p class="workflow-type">Type: Approval</p>
			<div class="workflow-state">
				<div class="state-indicator">
					<span class="state current">under_review</span>
					<span class="arrow">→</span>
					<span class="state">approved</span>
				</div>
			</div>
			<div class="workflow-meta">
				<span>Started: 30 minutes ago</span>
			</div>
		</div>

		<div class="workflow-card">
			<div class="workflow-header">
				<h3>ETL Pipeline</h3>
				<span class="status completed">Completed</span>
			</div>
			<p class="workflow-type">Type: Data Pipeline</p>
			<div class="workflow-state">
				<div class="state-indicator">
					<span class="state">extracting</span>
					<span class="arrow">→</span>
					<span class="state">transforming</span>
					<span class="arrow">→</span>
					<span class="state current">completed</span>
				</div>
			</div>
			<div class="workflow-meta">
				<span>Completed: 1 hour ago</span>
			</div>
		</div>
	</div>
</div>

<style>
	.workflows {
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

	.workflow-list {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: 1.5rem;
	}

	.workflow-card {
		background: white;
		padding: 1.5rem;
		border-radius: 8px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		transition: transform 0.2s, box-shadow 0.2s;
	}

	.workflow-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		cursor: pointer;
	}

	.workflow-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.workflow-header h3 {
		margin: 0;
		color: #1a1a1a;
	}

	.workflow-type {
		margin: 0 0 1rem 0;
		color: #666;
		font-size: 0.9rem;
	}

	.workflow-state {
		margin: 1rem 0;
		padding: 1rem;
		background-color: #f8f9fa;
		border-radius: 4px;
	}

	.state-indicator {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.85rem;
	}

	.state {
		padding: 0.25rem 0.5rem;
		background-color: #e9ecef;
		border-radius: 4px;
		color: #495057;
	}

	.state.current {
		background-color: #007bff;
		color: white;
		font-weight: 600;
	}

	.arrow {
		color: #adb5bd;
	}

	.workflow-meta {
		font-size: 0.85rem;
		color: #888;
	}

	.status {
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.status.running {
		background-color: #17a2b8;
		color: white;
	}

	.status.pending {
		background-color: #ffc107;
		color: #000;
	}

	.status.completed {
		background-color: #28a745;
		color: white;
	}
</style>
