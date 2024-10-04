document.addEventListener("DOMContentLoaded", (event) => {
    document.body.addEventListener("htmx:beforeSwap", (evt) => {
        if (evt.detail.xhr.status === 422 || evt.detail.xhr.status === 400) {
            const errorDialog = document.getElementById("errorDialog");
            const errorMessage = document.getElementById("errorMessage");
            errorMessage.textContent = evt.detail.xhr.response;
            errorDialog.showModal();
            evt.detail.shouldSwap = false;
        }
    });
});
