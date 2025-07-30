import { pgTable, text, timestamp, varchar, primaryKey } from 'drizzle-orm/pg-core';
import { relations } from 'drizzle-orm';
import { notices } from './notices';

export const keywords = pgTable('Keyword', {
  id: varchar('id').primaryKey().default('uuid_generate_v4()'),
  value: text('value').unique().notNull(),
  createdAt: timestamp('createdAt').defaultNow().notNull(),
  updatedAt: timestamp('updatedAt').defaultNow().notNull(),
});

export const keywordToNotice = pgTable('_KeywordToNotice', {
  keywordId: varchar('A').notNull(),
  noticeId: varchar('B').notNull(),
}, (table) => {
  return {
    pk: primaryKey({ columns: [table.keywordId, table.noticeId] }),
  };
});

export const keywordsRelations = relations(keywords, ({ many }) => ({
  notices: many(keywordToNotice),
}));

export const keywordToNoticeRelations = relations(keywordToNotice, ({ one }) => ({
  keyword: one(keywords, {
    fields: [keywordToNotice.keywordId],
    references: [keywords.id],
  }),
  notice: one(notices, {
    fields: [keywordToNotice.noticeId],
    references: [notices.id],
  }),
}));