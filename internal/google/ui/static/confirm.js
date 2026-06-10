// Confirm dialog handler using data-confirm attribute (XSS-safe replacement for inline onsubmit/hx-confirm)
(function() {
    // Handle standard form submissions
    document.addEventListener('submit', function(e) {
        var form = e.target.closest('form[data-confirm]');
        if (form && !confirm(form.getAttribute('data-confirm'))) {
            e.preventDefault();
            e.stopImmediatePropagation();
        }
    }, true);

    // Handle htmx:confirm events — ONLY when the element does NOT have an ancestor
    // form[data-confirm] (to avoid double-dialog with the submit handler above).
    document.addEventListener('htmx:confirm', function(e) {
        var elt = e.detail.elt;
        var msg = elt.getAttribute('data-confirm');
        if (!msg) return;
        // If the element or its ancestor is a form with data-confirm, the submit
        // handler already handles it — skip to avoid double dialogs.
        if (elt.closest && elt.closest('form[data-confirm]')) return;
        e.preventDefault();
        if (confirm(msg)) {
            e.detail.issueRequest(true);
        }
    });
})();
