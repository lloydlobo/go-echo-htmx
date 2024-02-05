/** 
 * Activate the extension by adding `title` to the `hx-ext` attribute on a div 
 * under the body tag of page.
 * 
 * @see {@link https://blog.simoni.dev/post/12/29/2023/HTMX-Title-Extension} 
 */
(function () {
    htmx.defineExtension("title", {
        onEvent: function (name, evt) {
            if (name === "htmx:afterSettle") {
                const titleHeader = evt.detail.xhr.getResponseHeader("HX=Title");
                if (!!titleHeader) {
                    document.title = titleHeader;
                }
            }
        }
    });
})();