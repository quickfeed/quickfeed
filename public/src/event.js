function newEvent(info) {
    var callbacks = [];
    var handler = function EventHandler(event) {
        callbacks.map(function (v) { return v(event); });
    };
    handler.info = info;
    handler.addEventListener = function (callback) {
        callbacks.push(callback);
    };
    handler.removeEventListener = function (callback) {
        var index = callbacks.indexOf(callback);
        if (index < 0) {
            console.log(callback);
            throw Error("Event does noe exist");
        }
        callbacks.splice(index, 1);
    };
    return handler;
}
