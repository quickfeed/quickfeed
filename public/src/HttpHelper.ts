import { combinePath } from "./NavigationHelper";

export interface IHTTPResult<T> {
    statusCode: number;
    data?: T;
}

/**
 * Wrapper around the buildt in XMLHttpRequest object
 */
export class HttpHelper {
    private PATH_PREFIX = "";

    get pathPrefix() {
        return this.PATH_PREFIX;
    }

    constructor(pathPrefix: string) {
        this.PATH_PREFIX = pathPrefix;
    }

    public get<T>(uri: string): Promise<IHTTPResult<T>> {
        return this.send("GET", uri);
    }

    public post<TSend, TReceive>(uri: string, sendData: TSend): Promise<IHTTPResult<TReceive>> {
        return this.send("POST", uri, sendData);
    }

    public put<TSend, TReceive>(uri: string, sendData: TSend): Promise<IHTTPResult<TReceive>> {
        return this.send("PUT", uri, sendData);
    }

    public delete<T>(uri: string): Promise<IHTTPResult<T>> {
        return this.send("DELETE", uri);
    }

    public patch<TSend, TReceive>(uri: string, sendData: TSend): Promise<IHTTPResult<TReceive>> {
        return this.send("PATCH", uri, sendData);
    }

    private send<TSend, TReceive>(method: string, uri: string, sendData?: TSend): Promise<IHTTPResult<TReceive>> {
        const request = new XMLHttpRequest();
        const requestPromise = new Promise<IHTTPResult<TReceive>>((resolve, reject) => {
            request.onreadystatechange = (ev: Event) => {
                if (request.readyState === 4) {
                    let data: TReceive | undefined;
                    const responseText = request.responseText.trim();
                    if (request.responseText.length < 2) {
                        console.log("Empty response detected");
                    } else if (responseText[0] !== "{" && responseText[0] !== "[") {
                        console.log("Non JSON respons detected");
                    } else {
                        try {
                            // console.log(request.responseText);
                            data = JSON.parse(request.responseText) as TReceive;
                        } catch (e) {
                            console.error("Could not parse response from server", e, request.responseText);
                        }
                    }
                    const temp: IHTTPResult<TReceive> = {
                        data,
                        statusCode: request.status,
                    };
                    resolve(temp);
                }
            };
            request.open(method, combinePath(this.PATH_PREFIX, uri), true);
            request.setRequestHeader("Content-Type", "application/json");
            if (sendData) {
                request.send(JSON.stringify(sendData));
            } else {
                request.send();
            }
        });
        return requestPromise;
    }
}
