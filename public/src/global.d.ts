declare module "*.scss"

declare const process: {
    env: {
        NODE_ENV: "development" | "production" | "test"
        QUICKFEED_ORGANIZATION_URL?: string
        QUICKFEED_APP_URL?: string
    }
}
