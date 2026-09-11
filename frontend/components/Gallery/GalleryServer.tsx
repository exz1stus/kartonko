"use server";
import { ImageMetadata } from "@/lib/api/generated/model";
import Gallery from "./Gallery";
import { SearchQuery } from "./ImageSearch";
import { getImage } from "@/lib/api/generated/server";
import ApiError from "@/lib/api/error";

interface Props {
    initialFetchSize?: number;
    initialQuery?: SearchQuery;
}

const INITIAL_QUERY: SearchQuery = {
    prefix: "",
    tags: [],
};

const GalleryServer = async ({
    initialFetchSize = 50,
    initialQuery = INITIAL_QUERY,
}: Props) => {
    const fetchImages = async (
        intialFetchSize: number,
        initialQuery: SearchQuery,
    ) => {
        const response = await getImage(
            {
                ...initialQuery,
                limit: intialFetchSize,
            },
            { cache: "no-store" },
        );

        console.log(
            `server intial fetching cursor ${0} limit ${intialFetchSize}`,
        );
        return response;
    };

    let images: ImageMetadata[] = [];
    let initReachedEnd = false;
    if (initialFetchSize > 0) {
        try {
            images = await fetchImages(initialFetchSize, initialQuery);
            initReachedEnd = images.length < initialFetchSize;
        } catch (error) {
            console.error("Error fetching images:", error);
            images = [];
        }
    }

    return (
        <Gallery
            initialImages={images}
            initReachedEnd={initReachedEnd}
            initialQuery={initialQuery}
        />
    );
};

export default GalleryServer;
