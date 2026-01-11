<script lang="ts">
	// Data lineage page
</script>

<div class="lineage">
	<div class="page-header">
		<h1>Data Lineage</h1>
		<button class="btn-primary">View Graph</button>
	</div>

	<div class="lineage-events">
		<div class="event-card">
			<div class="event-header">
				<span class="event-type start">START</span>
				<span class="event-time">2024-01-11 10:30:00 UTC</span>
			</div>
			<h3>Customer Data Pipeline</h3>
			<div class="event-details">
				<div class="detail-section">
					<h4>Job</h4>
					<p>namespace: opendq</p>
					<p>name: customer-etl-v1</p>
				</div>
				<div class="detail-section">
					<h4>Inputs</h4>
					<p>customers_raw (PostgreSQL)</p>
				</div>
				<div class="detail-section">
					<h4>Outputs</h4>
					<p>customers_processed (S3)</p>
				</div>
			</div>
		</div>

		<div class="event-card">
			<div class="event-header">
				<span class="event-type complete">COMPLETE</span>
				<span class="event-time">2024-01-11 10:32:15 UTC</span>
			</div>
			<h3>Customer Data Pipeline</h3>
			<div class="event-details">
				<div class="detail-section">
					<h4>Job</h4>
					<p>namespace: opendq</p>
					<p>name: customer-etl-v1</p>
				</div>
				<div class="detail-section">
					<h4>Duration</h4>
					<p>2 minutes 15 seconds</p>
				</div>
				<div class="detail-section">
					<h4>Records Processed</h4>
					<p>12,456 records</p>
				</div>
			</div>
		</div>

		<div class="event-card">
			<div class="event-header">
				<span class="event-type running">RUNNING</span>
				<span class="event-time">2024-01-11 10:35:00 UTC</span>
			</div>
			<h3>Analytics Aggregation</h3>
			<div class="event-details">
				<div class="detail-section">
					<h4>Job</h4>
					<p>namespace: opendq</p>
					<p>name: analytics-agg-v2</p>
				</div>
				<div class="detail-section">
					<h4>Inputs</h4>
					<p>customers_processed (S3)</p>
					<p>orders_data (PostgreSQL)</p>
				</div>
				<div class="detail-section">
					<h4>Outputs</h4>
					<p>customer_analytics (BigQuery)</p>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.lineage {
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

	.lineage-events {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.event-card {
		background: white;
		padding: 1.5rem;
		border-radius: 8px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		transition: transform 0.2s, box-shadow 0.2s;
	}

	.event-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.event-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.event-type {
		padding: 0.25rem 0.75rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.event-type.start {
		background-color: #17a2b8;
		color: white;
	}

	.event-type.complete {
		background-color: #28a745;
		color: white;
	}

	.event-type.running {
		background-color: #ffc107;
		color: #000;
	}

	.event-type.fail {
		background-color: #dc3545;
		color: white;
	}

	.event-time {
		color: #888;
		font-size: 0.85rem;
		font-family: monospace;
	}

	.event-card h3 {
		margin: 0 0 1rem 0;
		color: #1a1a1a;
	}

	.event-details {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.5rem;
		padding: 1rem;
		background-color: #f8f9fa;
		border-radius: 4px;
	}

	.detail-section h4 {
		margin: 0 0 0.5rem 0;
		color: #495057;
		font-size: 0.85rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.detail-section p {
		margin: 0.25rem 0;
		color: #666;
		font-size: 0.9rem;
		font-family: monospace;
	}
</style>
