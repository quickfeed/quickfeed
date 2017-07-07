export interface IHTTPResult<T> {
    statusCode: number;
    data: T;
}

export class HttpHelper {
    private PATH_PREFIX = "";
    get pathPrefix() {
        return this.PATH_PREFIX;
    }

    constructor(pathPrefix: string) {
        this.PATH_PREFIX = pathPrefix;
    }

    public get<T>(uri: string): Promise<IHTTPResult<T>> {
        return this.send("get", uri);
    }

    public post<TSend, TReceive>(uri: string, sendData: TSend): Promise<IHTTPResult<TReceive>> {
        return this.send("POST", uri, sendData);
    }

    private send<TSend, TReceive>(method: string, uri: string, sendData?: TSend): Promise<IHTTPResult<TReceive>> {
        const request = new XMLHttpRequest();
        const requestPromise = new Promise<IHTTPResult<TReceive>>((resolve, reject) => {
            request.onreadystatechange = (ev: Event) => {
                if (request.readyState === 4) {
                    const temp: IHTTPResult<TReceive> = {
                        data: JSON.parse(request.responseText) as TReceive,
                        statusCode: request.status,
                    };
                    resolve(temp);
                }
            };
            request.open(method, this.PATH_PREFIX + uri, true);
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
