"use server";
import { getUserMe } from "../api/generated/server";
import { UserDataResponse } from "../api/generated/model";
import ApiError from "../api/error";

export async function getLoggedUser(): Promise<UserDataResponse | null> {
    try {
        return await getUserMe({ credentials: "include" });
    } catch (error) {
        console.log("CAUGHT:", error);
        if (error instanceof ApiError && error.status === 401) {
            return null;
        }

        throw error;
    }
}
