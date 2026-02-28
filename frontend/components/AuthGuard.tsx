"use client";
import { useAuth } from "@/contexts/AuthContext";
import { useRouter } from "next/navigation";
import { ReactNode, useEffect } from "react";

interface Props {
    moderator?: boolean;
    children: ReactNode;
}

const AuthGuard = ({ moderator, children }: Props) => {
    const { user, loading } = useAuth();
    const router = useRouter();
    useEffect(() => {
        if (loading) return;
        if (!user) {
            router.replace(`/login?redirect=${encodeURIComponent(window.location.pathname)}`);
        }
    }, [user, loading, router]);

    if (!user) return <div>Loading...</div>;
    if(moderator && user.privileage !== "Moderator") return <div>Forbidden</div>;
    return <>{children}</>;
};

export default AuthGuard;
