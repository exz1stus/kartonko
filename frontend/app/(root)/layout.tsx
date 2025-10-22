import "@/app/globals.css";
import ClientShell from "../ClientShell";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en" className="h-screen">
            <ClientShell>
                <main className="h-full overflow-hidden">
                    {children}
                </main>
            </ClientShell>
        </html >
    );
}


