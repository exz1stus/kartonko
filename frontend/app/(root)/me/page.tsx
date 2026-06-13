"use client";
import AuthGuard from "@/components/AuthGuard";
import Loading from "@/components/Loading";
import { useAuth } from "@/contexts/AuthContext";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function MePage() {
    const { user, loading } = useAuth();
    const router = useRouter();
    useEffect(() => {
        if (loading || !user) return;
        else router.replace(`/user/${user?.username}`);
    }, [user, loading, router]);

    return (
        <AuthGuard>
            <Loading />
        </AuthGuard>
    );
}
