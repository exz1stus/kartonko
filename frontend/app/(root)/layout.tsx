import "@/app/globals.css";
import ClientShell from "@/app/ClientShell";

export const metadata = { title: "kartonko" };

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en" className="h-screen" suppressHydrationWarning>
            <body className="grid grid-rows-[auto_1fr] bg-image bg-surface-10 h-screen overflow-hidden">
                <ClientShell>
                    <main className="h-full overflow-hidden">{children}</main>
                </ClientShell>
            </body>
        </html>
    );
}
