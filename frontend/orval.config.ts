import { defineConfig } from "orval";

export default defineConfig({
    client: {
        input: {
            target: "./lib/api/openapi.json",
        },
        output: {
            target: "./lib/api/generated/client.ts",
            schemas: "./lib/api/generated/model",
            client: "fetch",
            override: {
                mutator: {
                    path: "./lib/api/clientMutator.ts",
                    name: "clientMutator",
                },
                fetch: {
                    includeHttpResponseReturnType: false,
                },
            },
        },
    },
    server: {
        input: {
            target: "./lib/api/openapi.json",
        },
        output: {
            target: "./lib/api/generated/server.ts",
            client: "fetch",
            override: {
                mutator: {
                    path: "./lib/api/serverMutator.ts",
                    name: "serverMutator",
                },
                fetch: {
                    includeHttpResponseReturnType: false,
                },
            },
        },
    },
});
