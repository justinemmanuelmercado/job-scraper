import { pgTable, text, timestamp, varchar } from 'drizzle-orm/pg-core';
import { relations } from 'drizzle-orm';
import { notices } from './notices';

export const sources = pgTable('Source', {
  name: varchar('name').primaryKey(),
  description: text('description').default(''),
  createdAt: timestamp('createdAt').defaultNow().notNull(),
  updatedAt: timestamp('updatedAt').defaultNow().notNull(),
  homepage: text('homepage').default(''),
  iconUrl: text('iconUrl').default(''),
});

export const sourcesRelations = relations(sources, ({ many }) => ({
  notices: many(notices),
}));