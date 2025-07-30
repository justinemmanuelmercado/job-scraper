import { pgTable, text, timestamp, varchar, unique } from 'drizzle-orm/pg-core';
import { relations } from 'drizzle-orm';
import { sources } from './sources';
import { keywordToNotice } from './keywords';

export const notices = pgTable('Notice', {
  id: varchar('id').primaryKey().default('uuid_generate_v4()'),
  title: text('title').notNull(),
  body: text('body').notNull(),
  url: text('url').notNull(),
  authorName: text('authorName').notNull(),
  authorUrl: text('authorUrl').notNull(),
  imageUrl: text('imageUrl'),
  createdAt: timestamp('createdAt').defaultNow().notNull(),
  updatedAt: timestamp('updatedAt'),
  sourceId: text('sourceId').notNull(),
  raw: text('raw').notNull(),
  guid: text('guid'),
  publishedDate: timestamp('publishedDate'),
}, (table) => {
  return {
    uniqueGuidSource: unique().on(table.guid, table.sourceId),
  };
});

export const noticesRelations = relations(notices, ({ one, many }) => ({
  source: one(sources, {
    fields: [notices.sourceId],
    references: [sources.name],
  }),
  keywords: many(keywordToNotice),
}));