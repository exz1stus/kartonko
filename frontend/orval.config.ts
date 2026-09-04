import { defineConfig } from "orval";

export default defineConfig({
    api: {
        input: {
            target: "./lib/openapi.json",
        },
        output: {
            target: "./lib/generated/api.ts",
            schemas: "./lib/generated/model",
            client: "fetch",
        },
    },
});
