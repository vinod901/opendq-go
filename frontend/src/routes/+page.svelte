<script lang="ts">
	// Dashboard page
</script>

<div class="dashboard">
	<h1>Dashboard</h1>
	
	<div class="stats-grid">
		<div class="stat-card">
			<h3>Tenants</h3>
			<div class="stat-value">12</div>
			<p class="stat-description">Active organizations</p>
		</div>
		
		<div class="stat-card">
			<h3>Policies</h3>
			<div class="stat-value">48</div>
			<p class="stat-description">Governance policies</p>
		</div>
		
		<div class="stat-card">
			<h3>Workflows</h3>
			<div class="stat-value">156</div>
			<p class="stat-description">Active workflows</p>
		</div>
		
		<div class="stat-card">
			<h3>Lineage Events</h3>
			<div class="stat-value">2.4K</div>
			<p class="stat-description">Data lineage records</p>
		</div>
	</div>

	<div class="features">
		<h2>Platform Features</h2>
		
		<div class="feature-grid">
			<div class="feature-card">
				<h3>🏢 Multi-Tenant</h3>
				<p>Isolated namespaces for multiple organizations with complete data separation</p>
			</div>
			
			<div class="feature-card">
				<h3>🔐 OpenFGA Authorization</h3>
				<p>Fine-grained access control using relationship-based authorization</p>
			</div>
			
			<div class="feature-card">
				<h3>🔑 OIDC Authentication</h3>
				<p>Enterprise SSO with Okta and Keycloak integration</p>
			</div>
			
			<div class="feature-card">
				<h3>📋 Policy-Driven</h3>
				<p>Flexible policy engine for data governance and compliance</p>
			</div>
			
			<div class="feature-card">
				<h3>🔄 Workflow Engine</h3>
				<p>State machine-based workflow orchestration with FSM</p>
			</div>
			
			<div class="feature-card">
				<h3>📊 Data Lineage</h3>
				<p>OpenLineage compatible tracking for end-to-end data visibility</p>
			</div>
		</div>
	</div>
</div>

<style>
	.dashboard {
		padding: 1rem;
	}

	h1 {
		color: #333;
		margin-bottom: 2rem;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1.5rem;
		margin-bottom: 3rem;
	}

	.stat-card {
		background: white;
		padding: 1.5rem;
		border-radius: 8px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}

	.stat-card h3 {
		margin: 0 0 1rem 0;
		color: #666;
		font-size: 0.9rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.stat-value {
		font-size: 2.5rem;
		font-weight: bold;
		color: #1a1a1a;
		margin-bottom: 0.5rem;
	}

	.stat-description {
		margin: 0;
		color: #888;
		font-size: 0.9rem;
	}

	.features {
		margin-top: 3rem;
	}

	.features h2 {
		color: #333;
		margin-bottom: 1.5rem;
	}

	.feature-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1.5rem;
	}

	.feature-card {
		background: white;
		padding: 1.5rem;
		border-radius: 8px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		transition: transform 0.2s, box-shadow 0.2s;
	}

	.feature-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.feature-card h3 {
		margin: 0 0 1rem 0;
		color: #1a1a1a;
		font-size: 1.2rem;
	}

	.feature-card p {
		margin: 0;
		color: #666;
		line-height: 1.6;
	}
</style>

