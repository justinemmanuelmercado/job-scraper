# humans.md

This file contains human-readable documentation for maintaining and developing the job scraper project.

## Database Maintenance

This project uses Drizzle ORM for database schema management. The database layer is managed through TypeScript/Node.js tooling while the main application is written in Go.

### Prerequisites

- Node.js and pnpm installed
- PostgreSQL running locally
- Environment variables configured in `.env` file

### Schema Management

#### Defining Schema

Database schema is defined in TypeScript files located in `db/schema/`:

- `db/schema/notices.ts` - Job posting table definition
- `db/schema/sources.ts` - Job source table definition  
- `db/schema/keywords.ts` - Keywords and keyword-to-notice junction table
- `db/schema/index.ts` - Exports all schema definitions

When modifying the schema:
1. Edit the appropriate schema file in `db/schema/`
2. Use Drizzle ORM column types and constraints
3. Update relations if adding/removing foreign keys
4. Export new tables from `db/schema/index.ts`

#### Creating Migrations

After modifying schema files, generate a new migration:

```bash
pnpm db:generate
```

This command:
- Reads schema files from `db/schema/`
- Compares with existing migrations
- Generates new SQL migration file in `db/migrations/`
- Names migrations automatically (e.g., `0001_brave_wolverine.sql`)

#### Applying Migrations

To apply pending migrations to your database:

```bash
pnpm db:migrate
```

This command:
- Connects to database using `DATABASE_URL` from `.env`
- Applies all pending migrations in order
- Updates migration tracking table

#### Development Workflow

For rapid development, you can push schema changes directly without creating migration files:

```bash
pnpm db:push
```

**Warning**: This bypasses migration history and should only be used in development.

#### Database Inspection

Launch Drizzle Studio to inspect your database:

```bash
pnpm db:studio
```

This opens a web interface at `https://local.drizzle.studio` for browsing tables and data.

#### Emergency Operations

To drop all tables (destructive operation):

```bash
pnpm db:drop
```

**Warning**: This will permanently delete all data. Use with extreme caution.

### Drizzle Kit Scripts

The following npm scripts are available in `package.json`:

- `pnpm db:generate` - Generate migration files from schema changes
- `pnpm db:migrate` - Apply pending migrations to database
- `pnpm db:push` - Push schema changes directly (development only)
- `pnpm db:studio` - Open Drizzle Studio for database inspection
- `pnpm db:drop` - Drop all tables (destructive)

### Configuration

Database configuration is managed in `drizzle.config.ts`:

```typescript
export default defineConfig({
  schema: './db/schema/*',           // Schema file location
  out: './db/migrations',            // Migration output directory
  dialect: 'postgresql',             // Database type
  dbCredentials: {
    url: process.env.DATABASE_URL!,  // Connection string from .env
  },
  verbose: true,                     // Detailed logging
  strict: true,                      // Strict mode for safety
});
```

### Environment Variables

Required environment variables for database operations:

```
DATABASE_URL=postgresql://username:password@localhost:5432/database_name
```

Alternative format using separate variables:
```
POSTGRES_USER=username
POSTGRES_PASSWORD=password  
POSTGRES_DB=database_name
```

### Best Practices

1. **Always generate migrations** for schema changes in production environments
2. **Test migrations** on a copy of production data before applying
3. **Review generated SQL** in migration files before applying
4. **Backup database** before running destructive operations
5. **Use `db:push`** only in development for rapid iteration
6. **Commit migration files** to version control along with schema changes
7. **Run migrations** as part of deployment process

### Troubleshooting

**Migration conflicts**: If multiple developers create migrations simultaneously, you may need to resolve conflicts by:
1. Pulling latest changes
2. Regenerating migrations if needed
3. Testing migration order

**Connection issues**: Verify:
- PostgreSQL is running
- Environment variables are correct
- Database exists and is accessible
- User has appropriate permissions

**Schema drift**: If database state doesn't match migrations:
1. Check migration history in database
2. Use `pnpm db:push` to sync development database
3. Regenerate migrations if necessary