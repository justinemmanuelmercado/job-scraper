CREATE TABLE "Notice" (
	"id" varchar PRIMARY KEY DEFAULT 'uuid_generate_v4()' NOT NULL,
	"title" text NOT NULL,
	"body" text NOT NULL,
	"url" text NOT NULL,
	"authorName" text NOT NULL,
	"authorUrl" text NOT NULL,
	"imageUrl" text,
	"createdAt" timestamp DEFAULT now() NOT NULL,
	"updatedAt" timestamp,
	"sourceId" text NOT NULL,
	"raw" text NOT NULL,
	"guid" text,
	"publishedDate" timestamp,
	CONSTRAINT "Notice_guid_sourceId_unique" UNIQUE("guid","sourceId")
);
--> statement-breakpoint
CREATE TABLE "Source" (
	"name" varchar PRIMARY KEY NOT NULL,
	"description" text DEFAULT '',
	"createdAt" timestamp DEFAULT now() NOT NULL,
	"updatedAt" timestamp DEFAULT now() NOT NULL,
	"homepage" text DEFAULT '',
	"iconUrl" text DEFAULT ''
);
--> statement-breakpoint
CREATE TABLE "_KeywordToNotice" (
	"A" varchar NOT NULL,
	"B" varchar NOT NULL,
	CONSTRAINT "_KeywordToNotice_A_B_pk" PRIMARY KEY("A","B")
);
--> statement-breakpoint
CREATE TABLE "Keyword" (
	"id" varchar PRIMARY KEY DEFAULT 'uuid_generate_v4()' NOT NULL,
	"value" text NOT NULL,
	"createdAt" timestamp DEFAULT now() NOT NULL,
	"updatedAt" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "Keyword_value_unique" UNIQUE("value")
);
