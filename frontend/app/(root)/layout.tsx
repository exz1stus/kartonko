import "@/app/globals.css";
import ClientShell from "@/app/ClientShell";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en" className="h-screen">
            <head>
                <title>Kartonko</title>
                <link rel="icon" href="@/favicon.ico" sizes="any"></link>
            </head>
            <ClientShell>
                <main className="h-full overflow-hidden">{children}</main>
            </ClientShell>
        </html>
    );
}
