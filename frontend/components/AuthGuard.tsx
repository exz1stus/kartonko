"use client";
import { useAuth } from "@/contexts/AuthContext";
import { useRouter } from "next/navigation";
import { ReactNode, useEffect } from "react";

interface AuthGuardProps {
    children: ReactNode;
}

const AuthGuard = ({ children }: AuthGuardProps) => {
    const { user, loading } = useAuth();
    const router = useRouter();
    useEffect(() => {
        if (loading) return;
        if (!user) {
            router.replace(`/login?redirect=${encodeURIComponent(window.location.pathname)}`);
        }
    }, [user, loading, router]);

    if (!user) return <div>Loading...</div>;

    return <>{children}</>;
};

export default AuthGuard;
