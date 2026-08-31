declare module "*.scss"

declare namespace NodeJS {
    interface ProcessEnv {
        NODE_ENV: "development" | "production" | "test"
        QUICKFEED_ORGANIZATION_URL?: string
        QUICKFEED_APP_URL?: string
    }
}
