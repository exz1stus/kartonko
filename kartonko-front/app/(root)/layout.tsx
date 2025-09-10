import "@/app/globals.css";
import { AuthProvider } from "../AuthContext";
import NavBar from "@/components/Navbar/NavBar";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en" className="min-h-screen">
            <AuthProvider>
                <body className="min-h-screen">
                    <NavBar />
                    <div className="h-full flex flex-col items-center">
                        <main className="flex-1 h-full">
                            {children}
                        </main>
                    </div>
                </body>
            </AuthProvider>
        </html>
    );
}


