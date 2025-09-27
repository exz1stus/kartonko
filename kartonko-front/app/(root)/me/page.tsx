"use client";
import { useAuth } from "@/app/AuthContext"
import { redirect } from "next/navigation"

export default function MePage() {
  const username = useAuth().user?.username;

  if (!username) {
    redirect("/login")
  }

  redirect(`/user/${username}`)
}