import "@/app/globals.css";
import { AuthProvider } from "../AuthContext";
import NavBar from "@/components/NavBar";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en" className="h-full">
            <body className="h-full">
                <AuthProvider>
                    <div className="h-full flex flex-col items-center">
                        <NavBar />
                        <main className="flex-1 h-full">
                            {children}
                        </main>
                    </div>
                </AuthProvider>
            </body>
        </html>
    );
}


