export interface SearchQuery {
    prefix?: string;
    withTags?: string[];
    userID?: number;
}

export function constructQueryString(query: SearchQuery) {
    const prefixQueryString = query.prefix ? `prefix=${query.prefix}&` : "";
    const tagsQueryString =
        query.withTags && query.withTags.length > 0
            ? `tags=${JSON.stringify(query.withTags)}&`
            : "";
    const userID = query.userID ? `user_id=${query.userID}&` : "";

    let queryString = `${prefixQueryString}${tagsQueryString}${userID}`;

    if (queryString.length > 0)
        queryString = queryString.substring(0, queryString.length - 1);

    return queryString;
}
