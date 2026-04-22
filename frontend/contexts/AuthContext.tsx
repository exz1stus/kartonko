"use client";
import { usePathname } from "next/navigation";
import { useRouter } from "next/navigation";
import router from "next/router";
import {
    useState,
    useEffect,
    useCallback,
    createContext,
    useContext,
} from "react";
import { UserData } from "@/lib/user";
import { apiFetch } from "@/lib/apiFetch";

interface AuthContextType {
    user: UserData | null;
    login: (redirectPath?: string) => void;
    logout: () => void;
    loading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<React.PropsWithChildren<{}>> = ({
    children,
}) => {
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
    const router = useRouter();
    const pathname = usePathname();

    const fetchUser = useCallback(async () => {
        setLoading(true);
        try {
            const res = await apiFetch(`/me`, { credentials: "include" });
            if (res.status === 401 || res.status === 403) {
                setUser(null);
                return;
            }

            if (!res.ok) {
                throw new Error(`Request failed: ${res.status}`);
            }

            const user = await res.json();
            if (user?.username) {
                setUser(user);
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

    const login = (redirect?: string) => {
        const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;
        const FRONTEND_URI = process.env.NEXT_PUBLIC_FRONTEND_URI;
        const redirectPath = redirect || pathname || "/";

        window.location.href = `${API_ORIGIN}/auth/google?redirect=${encodeURIComponent(
            `${FRONTEND_URI}${redirectPath}`,
        )}`;
    };

    const logout = async () => {
        await apiFetch("/auth/logout", { method: "POST" });
        setUser(null);
        router.refresh();
    };

    return { user, loading, login, logout };
};
