(function () {
  "use strict";

  var searchForm = document.getElementById("search-form");
  var searchInput = document.getElementById("search-input");
  var searchBtn = document.getElementById("search-btn");
  var answerArea = document.getElementById("answer-area");
  var answerLoading = document.getElementById("answer-loading");
  var answerContent = document.getElementById("answer-content");
  var answerCitations = document.getElementById("answer-citations");
  var adrListItems = document.getElementById("adr-list-items");
  var aboutContent = document.getElementById("about-content");
  var adrDetail = document.getElementById("adr-detail");
  var headerTitle = document.getElementById("header-title");

  // --- ADR List (US2) ---

  function loadADRList() {
    fetch("/adrs")
      .then(function (res) { return res.json(); })
      .then(function (data) {
        renderADRList(data.adrs || []);
      })
      .catch(function () {
        adrListItems.innerHTML = '<li class="adr-list-empty">Failed to load ADRs</li>';
      });
  }

  function renderADRList(adrs) {
    if (adrs.length === 0) {
      adrListItems.innerHTML = '<li class="adr-list-empty">No ADRs available</li>';
      return;
    }
    adrListItems.innerHTML = "";
    adrs.forEach(function (adr) {
      var li = document.createElement("li");
      li.setAttribute("data-number", adr.number);
      li.innerHTML =
        '<span class="adr-list-number">ADR-' + String(adr.number).padStart(3, "0") + '</span>' +
        '<span class="adr-list-title">' + escapeHTML(adr.title) + '</span>' +
        '<span class="adr-list-status">' + escapeHTML(adr.status) + '</span>';
      li.addEventListener("click", function () {
        showADR(adr.number);
      });
      adrListItems.appendChild(li);
    });
  }

  // --- ADR Detail ---

  function showADR(number) {
    fetch("/adrs/" + number)
      .then(function (res) {
        if (!res.ok) throw new Error("ADR not found");
        return res.json();
      })
      .then(function (data) {
        adrDetail.innerHTML = marked.parse(data.content || "");
        aboutContent.classList.add("hidden");
        adrDetail.classList.remove("hidden");
        setActiveADR(number);
      })
      .catch(function () {
        adrDetail.innerHTML = '<p class="error-msg">Failed to load ADR</p>';
        aboutContent.classList.add("hidden");
        adrDetail.classList.remove("hidden");
      });
  }

  function setActiveADR(number) {
    var items = adrListItems.querySelectorAll("li");
    items.forEach(function (li) {
      if (parseInt(li.getAttribute("data-number"), 10) === number) {
        li.classList.add("active");
      } else {
        li.classList.remove("active");
      }
    });
  }

  // --- About Panel (US4) ---

  function showAbout() {
    aboutContent.classList.remove("hidden");
    adrDetail.classList.add("hidden");
    answerArea.classList.add("hidden");
    var items = adrListItems.querySelectorAll("li");
    items.forEach(function (li) { li.classList.remove("active"); });
  }

  headerTitle.addEventListener("click", showAbout);

  // --- Query (US1) ---

  searchForm.addEventListener("submit", function (e) {
    e.preventDefault();

    var query = searchInput.value.trim();
    if (!query) return;

    searchBtn.disabled = true;
    searchInput.disabled = true;

    answerArea.classList.remove("hidden");
    answerLoading.classList.remove("hidden");
    answerContent.innerHTML = "";
    answerCitations.innerHTML = "";

    fetch("/query", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ query: query })
    })
      .then(function (res) { return res.json(); })
      .then(function (data) {
        answerLoading.classList.add("hidden");

        if (data.error) {
          answerContent.innerHTML = '<p class="error-msg">' + escapeHTML(data.error) + '</p>';
          return;
        }

        answerContent.innerHTML = marked.parse(data.answer || "No relevant information found.");

        if (data.citations && data.citations.length > 0) {
          data.citations.forEach(function (c) {
            var link = document.createElement("a");
            link.className = "citation-link";
            link.textContent = "ADR-" + String(c.adr_number).padStart(3, "0");
            link.href = "#";
            link.addEventListener("click", function (ev) {
              ev.preventDefault();
              showADR(c.adr_number);
            });
            answerCitations.appendChild(link);
          });
        }
      })
      .catch(function () {
        answerLoading.classList.add("hidden");
        answerContent.innerHTML = '<p class="error-msg">Failed to get a response. Please try again.</p>';
      })
      .finally(function () {
        searchBtn.disabled = false;
        searchInput.disabled = false;
      });
  });

  // --- Utility ---

  function escapeHTML(str) {
    var div = document.createElement("div");
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  // --- Sample Queries ---

  var queryItems = document.querySelectorAll(".about-queries li");
  queryItems.forEach(function (li) {
    li.addEventListener("click", function () {
      searchInput.value = li.textContent.replace(/[\u201c\u201d]/g, "");
      searchForm.dispatchEvent(new Event("submit"));
    });
  });

  // --- Init ---

  loadADRList();
})();
