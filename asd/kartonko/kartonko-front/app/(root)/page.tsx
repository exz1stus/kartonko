import PostImageManager from "@/components/PostImageManager";
import NavBar from "@/components/NavBar";
import Gallery from "@/components/Gallery";

export default function Home() {
    return (
        <>
            <NavBar />
            <main>
                <Gallery/>
                <PostImageManager />
            </main>
        </>
    );
}
