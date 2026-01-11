# OpenDQ Frontend

SvelteKit-based frontend for the OpenDQ Control Plane Platform.

## Features

- **Dashboard**: Overview of tenants, policies, workflows, and lineage events
- **Tenant Management**: Manage multi-tenant organizations
- **Policy Management**: Configure governance and compliance policies
- **Workflow Management**: Monitor and control workflow states
- **Data Lineage**: View OpenLineage-compatible data lineage events

## Development

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```sh
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

## Building

To create a production version of your app:

```sh
npm run build
```

You can preview the production build with `npm run preview`.

## Configuration

The frontend can be configured to connect to the backend API:

Create a `.env` file:

```
VITE_API_URL=http://localhost:8080
```

## Architecture

- **SvelteKit**: Modern framework for building web applications
- **TypeScript**: Type-safe development
- **Vite**: Fast build tool and dev server

## Pages

- `/` - Dashboard
- `/tenants` - Tenant management
- `/policies` - Policy management
- `/workflows` - Workflow management
- `/lineage` - Data lineage viewer

## Future Enhancements

- [ ] API integration with backend
- [ ] Authentication flow (OIDC)
- [ ] Real-time updates via WebSockets
- [ ] Advanced lineage graph visualization
- [ ] Policy editor with validation
- [ ] Workflow transition controls
- [ ] User management
- [ ] Tenant creation wizard

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.

