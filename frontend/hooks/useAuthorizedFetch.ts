import { useCallback } from "react";
import { useAuth } from "../contexts/AuthContext";

const useAuthFetch = () => {
    const auth = useAuth();

    const authFetch = useCallback(async (input: RequestInfo, init?: RequestInit) => {
        try {
            const res = await fetch(input, {
                ...init,
                credentials: "include",
            });

            if (res.status === 401 || res.status === 403) {
                auth.login();
                return Promise.reject(new Error("Unauthorized"));
            }

            return res;
        } catch (err) {
            throw err;
        }
    }, []);

    return authFetch;
};

export default useAuthFetch;
