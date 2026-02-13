"use client";
import { Timestamp } from "next/dist/server/lib/cache-handlers/types";
import { usePathname } from "next/navigation";
import router from "next/router";
import { useState, useEffect, useCallback, createContext, useContext } from "react";
import { UserData } from "@/app/lib/user";

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;
const FRONTEND_URI = process.env.NEXT_PUBLIC_FRONTEND_URI;

const AuthContext = createContext<
    { user: UserData | null; login: Function; logout: Function; loading: boolean } | undefined
>(undefined);

export const AuthProvider: React.FC<React.PropsWithChildren<{}>> = ({ children }) => {
    const auth = useProvideAuth();

    useEffect(() => {
        const pageLeave = async () => {
            //unsaved changes
        };

        const handleWindowClose = (e: BeforeUnloadEvent) => {
            pageLeave();
        };
        const handleBrowseAway = () => {
            pageLeave();
        };

        window.addEventListener("beforeunload", handleWindowClose);
        router.events.on("routeChangeStart", handleBrowseAway);
        return () => {
            window.removeEventListener("beforeunload", handleWindowClose);
            router.events.off("routeChangeStart", handleBrowseAway);
        };
    }, []);

    return <AuthContext.Provider value={auth}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) throw new Error("useAuth must be inside AuthProvider");
    return context;
};

const useProvideAuth = () => {
    const [user, setUser] = useState<UserData | null>(null);
    const [loading, setLoading] = useState(true);

    const pathname = usePathname();

    const fetchUser = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetch(`${API_ORIGIN}/me`, { credentials: "include" });
            if (res.status === 401 || res.status === 403) {
                setUser(null);
                return;
            }

            if (!res.ok) {
                throw new Error(`Request failed: ${res.status}`);
            }

            const data = await res.json();
            if (data?.username && data?.picture_url) {
                setUser(data);
            } else {
                setUser(null);
            }
        } catch (err) {
            console.error("Failed to fetch user:", err);
            setUser(null);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchUser();
    }, [fetchUser, pathname]);

    const login = () => {
        const currentPath = pathname || "/";
        window.location.href = `${API_ORIGIN}/auth/google?redirect=${encodeURIComponent(
            `${FRONTEND_URI}${currentPath}`,
        )}`;
    };

    const logout = async () => {
        try {
            await fetch(`${API_ORIGIN}/auth/logout`, { credentials: "include" });
        } catch (err) {
            console.error("Logout failed:", err);
        }
        setUser(null);
        console.log("logged out", user);
    };

    return { user, loading, login, logout };
};

export { type UserData };
