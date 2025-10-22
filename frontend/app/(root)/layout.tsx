import "@/app/globals.css";
import { AuthProvider } from "../contexts/AuthContext";
import NavBar from "@/components/Navbar/NavBar";
import SideBarProvider from "../contexts/SidebarContext";
import ResizableGrid from "@/components/ResizableGrid";
import HoverProvider from "../contexts/HoverContex";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en" className="h-screen">
            <AuthProvider>
                <HoverProvider>
                    <SideBarProvider>
                        <body className="grid grid-rows-[auto_1fr] bg-image bg-surface-10 h-screen overflow-hidden"
                            style={{
                                backgroundImage: `url(http://localhost:8080/raw-image/rem)`,
                            }}
                        >
                            <NavBar />
                            <ResizableGrid>
                                <main className="h-full overflow-hidden">
                                    {children}
                                </main>
                            </ResizableGrid>
                        </body>
                    </SideBarProvider>
                </HoverProvider>
            </AuthProvider>
        </html >
    );
}


