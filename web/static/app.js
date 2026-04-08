(function () {
  "use strict";

  // --- Utility ---

  function escapeHTML(str) {
    var div = document.createElement("div");
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  function linkifyADRReferences(html) {
    var knownNumbers = {};
    Alpine.store("adrs").all.forEach(function (a) { knownNumbers[a.number] = true; });
    return html.replace(/ADR-(\d+)/g, function (match, numStr) {
      var num = parseInt(numStr, 10);
      if (knownNumbers[num]) {
        return '<a href="#" class="citation-inline" data-adr="' + num + '">' + match + '</a>';
      }
      return match;
    });
  }

  function renderRelationships(rels) {
    var groups = {};
    var typeLabels = {
      supersedes: "Supersedes",
      superseded_by: "Superseded By",
      depends_on: "Depends On",
      drives: "Drives",
      related_to: "Related"
    };
    rels.forEach(function (r) {
      var label = typeLabels[r.rel_type] || r.rel_type;
      if (!groups[label]) groups[label] = [];
      groups[label].push(r);
    });
    var html = '<div class="adr-relationships"><h3>Related ADRs</h3>';
    Object.keys(groups).forEach(function (label) {
      html += "<h4>" + escapeHTML(label) + "</h4><ul>";
      groups[label].forEach(function (r) {
        html += '<li><a href="#" class="rel-link" data-adr="' + r.target_adr +
          '">ADR-' + String(r.target_adr).padStart(3, "0") + ": " +
          escapeHTML(r.target_title || "Unknown") + "</a></li>";
      });
      html += "</ul>";
    });
    html += "</div>";
    return html;
  }

  // --- Normalize status for badge CSS class ---

  window.normalizeStatus = function (status) {
    if (!status) return "unknown";
    var first = status.split(/[\s(]/)[0].toLowerCase();
    if (["accepted", "proposed", "deprecated", "superseded"].indexOf(first) >= 0) return first;
    return "unknown";
  };

  // --- Alpine Stores ---

  document.addEventListener("alpine:init", function () {

    // Navigation store
    Alpine.store("nav", {
      stack: [],

      get currentView() {
        return this.stack.length > 0 ? this.stack[this.stack.length - 1] : null;
      },

      get hasAnswer() {
        return this.stack.some(function (e) { return e.type === "answer"; });
      },

      get breadcrumbItems() {
        var items = [];
        var max = 5;
        var stack = this.stack;
        if (stack.length <= max) {
          stack.forEach(function (entry, idx) {
            items.push({ label: entry.label, stackIndex: idx });
          });
        } else {
          items.push({ label: stack[0].label, stackIndex: 0 });
          items.push({ label: "\u2026", stackIndex: -1 });
          for (var i = stack.length - (max - 2); i < stack.length; i++) {
            items.push({ label: stack[i].label, stackIndex: i });
          }
        }
        return items;
      },

      push: function (entry) {
        // If this ADR is already on the stack, go back to it instead of duplicating
        for (var i = 0; i < this.stack.length; i++) {
          if (this.stack[i].type === entry.type && entry.type === "adr" &&
              this.stack[i].data.number === entry.data.number) {
            this.stack = this.stack.slice(0, i + 1);
            return;
          }
        }
        this.stack = this.stack.concat([entry]);
      },

      replace: function (entry) {
        this.stack = [entry];
      },

      goto: function (index) {
        if (index < 0) return;
        if (index < this.stack.length) {
          this.stack = this.stack.slice(0, index + 1);
        }
      },

      clear: function () {
        this.stack = [];
      },

      goHome: function () {
        this.stack = [];
        Alpine.store("query").text = "";
        Alpine.store("query").error = null;
      },

      openADR: function (number) {
        var self = this;
        fetch("/adrs/" + number)
          .then(function (res) {
            if (!res.ok) throw new Error("ADR not found");
            return res.json();
          })
          .then(function (data) {
            var html = marked.parse(data.content || "");
            if (data.relationships && data.relationships.length > 0) {
              html += renderRelationships(data.relationships);
            }
            var entry = {
              type: "adr",
              label: "ADR-" + String(number).padStart(3, "0"),
              data: {
                number: number,
                title: data.title,
                status: data.status,
                date: data.date || "",
                contentHtml: html,
                relationships: data.relationships || []
              }
            };
            if (self.hasAnswer) {
              self.push(entry);
            } else {
              self.replace(entry);
            }
            if (window.innerWidth <= 768) {
              Alpine.store("ui").sidebarOpen = false;
            }
          })
          .catch(function () {
            var entry = {
              type: "adr",
              label: "ADR-" + String(number).padStart(3, "0"),
              data: {
                number: number,
                title: "Error",
                status: "",
                date: "",
                contentHtml: '<p class="error-msg">Failed to load ADR</p>',
                relationships: []
              }
            };
            if (self.hasAnswer) {
              self.push(entry);
            } else {
              self.replace(entry);
            }
          });
      }
    });

    // ADR list store
    Alpine.store("adrs", {
      all: [],
      filter: null,
      sort: "number-asc",
      loading: true,
      error: null,

      get filtered() {
        var self = this;
        var list = this.all;
        if (this.filter) {
          list = list.filter(function (adr) {
            return normalizeStatus(adr.status) === self.filter;
          });
        }
        var sorted = list.slice();
        switch (this.sort) {
          case "number-desc":
            sorted.sort(function (a, b) { return b.number - a.number; });
            break;
          case "date-desc":
            sorted.sort(function (a, b) { return (b.date || "").localeCompare(a.date || "") || b.number - a.number; });
            break;
          case "date-asc":
            sorted.sort(function (a, b) { return (a.date || "").localeCompare(b.date || "") || a.number - b.number; });
            break;
          default:
            sorted.sort(function (a, b) { return a.number - b.number; });
        }
        return sorted;
      },

      setFilter: function (f) {
        this.filter = f;
        try { localStorage.setItem("adr-insight-filter", JSON.stringify(f)); } catch (e) {}
      },

      setSort: function (s) {
        this.sort = s;
        try { localStorage.setItem("adr-insight-sort", s); } catch (e) {}
      },

      load: function () {
        var self = this;
        self.loading = true;
        self.error = null;
        fetch("/adrs")
          .then(function (res) { return res.json(); })
          .then(function (data) {
            self.all = data.adrs || [];
            self.loading = false;
          })
          .catch(function () {
            self.error = "Failed to load ADRs";
            self.loading = false;
          });
      },

      init: function () {
        try {
          var f = localStorage.getItem("adr-insight-filter");
          if (f) this.filter = JSON.parse(f);
        } catch (e) {}
        try {
          var s = localStorage.getItem("adr-insight-sort");
          if (s) this.sort = s;
        } catch (e) {}
        this.load();
      }
    });

    // Query store
    Alpine.store("query", {
      text: "",
      loading: false,
      loadingText: "Thinking...",
      error: null,
      lastQuery: "",
      history: [],
      showHistory: false,
      _thinkingTimer: null,

      submit: function () {
        var q = this.text.trim();
        if (!q) return;
        this.submitQuery(q);
      },

      submitQuery: function (q) {
        this.text = q;
        this.lastQuery = q;
        this.loading = true;
        this.loadingText = "Thinking...";
        this.error = null;
        this.showHistory = false;
        this.addToHistory(q);

        var nav = Alpine.store("nav");
        nav.stack = [{
          type: "answer",
          label: "Answer",
          data: { query: q, answerHtml: "", citations: [] }
        }];

        if (window.innerWidth <= 768) {
          Alpine.store("ui").sidebarOpen = false;
        }

        var self = this;
        if (this._thinkingTimer) clearTimeout(this._thinkingTimer);
        this._thinkingTimer = setTimeout(function () {
          if (self.loading) self.loadingText = "Still thinking...";
        }, 5000);

        fetch("/query", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ query: q })
        })
          .then(function (res) { return res.json(); })
          .then(function (data) {
            if (self._thinkingTimer) { clearTimeout(self._thinkingTimer); self._thinkingTimer = null; }
            self.loading = false;

            if (data.error) {
              self.error = data.error;
              return;
            }

            var answerHtml = marked.parse(data.answer || "No relevant information found.");
            answerHtml = linkifyADRReferences(answerHtml);

            nav.stack = [{
              type: "answer",
              label: "Answer",
              data: { query: q, answerHtml: answerHtml, citations: data.citations || [] }
            }];
          })
          .catch(function () {
            if (self._thinkingTimer) { clearTimeout(self._thinkingTimer); self._thinkingTimer = null; }
            self.loading = false;
            self.error = "Failed to get a response. Please try again.";
          });
      },

      retry: function () {
        if (this.lastQuery) {
          this.submitQuery(this.lastQuery);
        }
      },

      clearSearch: function () {
        this.text = "";
        this.error = null;
        this.showHistory = false;
        if (this._thinkingTimer) { clearTimeout(this._thinkingTimer); this._thinkingTimer = null; }
        Alpine.store("nav").clear();
      },

      selectHistory: function (q) {
        this.showHistory = false;
        this.submitQuery(q);
      },

      addToHistory: function (q) {
        this.history = [q].concat(this.history.filter(function (h) { return h !== q; })).slice(0, 10);
        try { localStorage.setItem("adr-insight-query-history", JSON.stringify(this.history)); } catch (e) {}
      },

      clearHistory: function () {
        this.history = [];
        this.showHistory = false;
        try { localStorage.removeItem("adr-insight-query-history"); } catch (e) {}
      },

      init: function () {
        try {
          var h = localStorage.getItem("adr-insight-query-history");
          if (h) this.history = JSON.parse(h);
        } catch (e) {}
      }
    });

    // UI store
    Alpine.store("ui", {
      sidebarOpen: window.innerWidth > 768
    });
  });

  // --- Delegated click handlers for dynamic content ---

  document.addEventListener("click", function (e) {
    var citLink = e.target.closest(".citation-inline");
    if (citLink) {
      e.preventDefault();
      var adrNum = parseInt(citLink.getAttribute("data-adr"), 10);
      if (adrNum) Alpine.store("nav").openADR(adrNum);
      return;
    }

    var relLink = e.target.closest(".rel-link");
    if (relLink) {
      e.preventDefault();
      var num = parseInt(relLink.getAttribute("data-adr"), 10);
      if (num) {
        var nav = Alpine.store("nav");
        // Relationship links always push onto the stack for chain navigation
        fetch("/adrs/" + num)
          .then(function (res) {
            if (!res.ok) throw new Error("ADR not found");
            return res.json();
          })
          .then(function (data) {
            var html = marked.parse(data.content || "");
            if (data.relationships && data.relationships.length > 0) {
              html += renderRelationships(data.relationships);
            }
            nav.push({
              type: "adr",
              label: "ADR-" + String(num).padStart(3, "0"),
              data: {
                number: num,
                title: data.title,
                status: data.status,
                date: data.date || "",
                contentHtml: html,
                relationships: data.relationships || []
              }
            });
          })
          .catch(function () {});
      }
      return;
    }
  });

})();
