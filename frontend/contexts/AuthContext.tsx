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
import { UserDataResponse } from "@/lib/api/generated/model/userDataResponse";
import { getUserMe, postAuthLogout } from "@/lib/api/generated/client";
import ApiError from "@/lib/api/error";

interface AuthContextType {
    user: UserDataResponse | null;
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
    const [user, setUser] = useState<UserDataResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const router = useRouter();
    const pathname = usePathname();

    const fetchUser = useCallback(async () => {
        setLoading(true);
        try {
            const user = await getUserMe({ credentials: "include" });
            setUser(user);
        } catch (error) {
            if (error instanceof ApiError && error.status === 401) {
                setUser(null);
                return;
            }

            throw error;
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
        await postAuthLogout();
        setUser(null);
        router.refresh();
    };

    return { user, loading, login, logout };
};
