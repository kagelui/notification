CREATE TABLE "public"."merchants"
(
    id           SERIAL PRIMARY KEY,
    business_id  TEXT UNIQUE              NOT NULL CHECK (business_id::TEXT <> ''::TEXT),
    token        TEXT                     NOT NULL CHECK (token::TEXT <> ''::TEXT),
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX merchants_business_id_index ON public.merchants (business_id);

CREATE TABLE "public"."callback_urls"
(
    id           SERIAL PRIMARY KEY,
    business_id  TEXT UNIQUE              NOT NULL CHECK (business_id::TEXT <> ''::TEXT),
    product_id   TEXT                     NOT NULL CHECK (product_id::TEXT <> ''::TEXT),
    callback_url TEXT                     NOT NULL CHECK (callback_url::TEXT <> ''::TEXT),
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX callback_urls_business_id_product_id_index ON public.callback_urls (business_id, product_id);

CREATE TABLE "public"."messages"
(
    id                 UUID PRIMARY KEY,
    product_id         TEXT                     NOT NULL CHECK (product_id::TEXT <> ''::TEXT),
    product_type       TEXT                     NOT NULL CHECK (product_type::TEXT <> ''::TEXT),
    payload            JSONB                    NOT NULL,
    merchant_id        INTEGER                  NOT NULL REFERENCES public.merchants (id),
    retry_count        INTEGER                  NOT NULL CHECK (retry_count >= 0),
    next_delivery_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status             TEXT                     NOT NULL CHECK (product_type::TEXT <> ''::TEXT),
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX messages_merchant_id_index ON public.messages (merchant_id);
