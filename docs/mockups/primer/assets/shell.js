/*
 * Shared chrome for the kfleet Primer mockups: octicon set, sidebar, topbar
 * and the light/dark theme switch. Pages declare their active nav item with
 * <body data-page="fleet"> and their breadcrumb with data-breadcrumb.
 */
(function () {
  var icons = {
    fleet:
      '<path d="M1.75 1h12.5c.966 0 1.75.784 1.75 1.75v9.5A1.75 1.75 0 0 1 14.25 14H1.75A1.75 1.75 0 0 1 0 12.25v-9.5C0 1.784.784 1 1.75 1ZM1.5 2.75v9.5c0 .138.112.25.25.25H7.5v-10H1.75a.25.25 0 0 0-.25.25Zm7.5-.25v10h5.25a.25.25 0 0 0 .25-.25v-9.5a.25.25 0 0 0-.25-.25Z"></path>',
    shield:
      '<path d="M7.467.133a1.748 1.748 0 0 1 1.066 0l5.25 1.68A1.75 1.75 0 0 1 15 3.48V7c0 1.566-.32 3.182-1.303 4.682-.983 1.498-2.585 2.813-5.032 3.855a1.697 1.697 0 0 1-1.33 0c-2.447-1.042-4.049-2.357-5.032-3.855C1.32 10.182 1 8.566 1 7V3.48a1.75 1.75 0 0 1 1.217-1.667Z"></path>',
    bell:
      '<path d="M8 16a2 2 0 0 0 1.985-1.75c.017-.137-.097-.25-.235-.25h-3.5c-.138 0-.252.113-.235.25A2 2 0 0 0 8 16ZM3 5a5 5 0 0 1 10 0v2.947c0 .05.015.098.042.139l1.703 2.555A1.519 1.519 0 0 1 13.482 13H2.518a1.516 1.516 0 0 1-1.263-2.359l1.703-2.555A.255.255 0 0 0 3 7.947Z"></path>',
    people:
      '<path d="M2 5.5a3.5 3.5 0 1 1 5.898 2.549 5.508 5.508 0 0 1 3.034 4.084.75.75 0 1 1-1.482.235 4 4 0 0 0-7.9 0 .75.75 0 0 1-1.482-.236A5.507 5.507 0 0 1 3.102 8.05 3.493 3.493 0 0 1 2 5.5ZM11 4a3.001 3.001 0 0 1 2.22 5.018 5.01 5.01 0 0 1 2.56 3.012.749.749 0 0 1-.885.954.752.752 0 0 1-.549-.514 3.507 3.507 0 0 0-2.522-2.372.75.75 0 0 1-.574-.73v-.352a.75.75 0 0 1 .416-.672A1.5 1.5 0 0 0 11 5.5.75.75 0 0 1 11 4Z"></path>',
    person:
      '<path d="M10.561 8.073a6.005 6.005 0 0 1 3.432 5.142.75.75 0 1 1-1.498.07 4.5 4.5 0 0 0-8.99 0 .75.75 0 0 1-1.498-.07 6.004 6.004 0 0 1 3.431-5.142A3.999 3.999 0 1 1 10.561 8.073ZM8 1.5a2.5 2.5 0 1 0 0 5 2.5 2.5 0 0 0 0-5Z"></path>',
    log:
      '<path d="M0 3.75A.75.75 0 0 1 .75 3h7.5a.75.75 0 0 1 0 1.5H.75A.75.75 0 0 1 0 3.75Zm0 4A.75.75 0 0 1 .75 7h12.5a.75.75 0 0 1 0 1.5H.75A.75.75 0 0 1 0 7.75Zm0 4a.75.75 0 0 1 .75-.75h9.5a.75.75 0 0 1 0 1.5H.75a.75.75 0 0 1-.75-.75Z"></path>',
    box:
      '<path d="M8.878.392a1.75 1.75 0 0 0-1.756 0l-5.25 3.045A1.75 1.75 0 0 0 1 4.951v6.098c0 .624.332 1.2.872 1.514l5.25 3.045a1.75 1.75 0 0 0 1.756 0l5.25-3.045c.54-.313.872-.89.872-1.514V4.951c0-.624-.332-1.2-.872-1.514ZM8 1.69l4.63 2.685L8 7.06 3.37 4.375Zm-5.5 3.98L7.25 8.43v5.24l-4.629-2.686a.25.25 0 0 1-.121-.216Zm6.25 7.999V8.43l4.75-2.759v5.096a.25.25 0 0 1-.121.216Z"></path>',
    server:
      '<path d="M1.75 1h12.5c.966 0 1.75.784 1.75 1.75v4c0 .966-.784 1.75-1.75 1.75H1.75A1.75 1.75 0 0 1 0 6.75v-4C0 1.784.784 1 1.75 1ZM1.5 2.75v4c0 .138.112.25.25.25h12.5a.25.25 0 0 0 .25-.25v-4a.25.25 0 0 0-.25-.25H1.75a.25.25 0 0 0-.25.25Zm0 6.75h12.5c.966 0 1.75.784 1.75 1.75v2.5A1.75 1.75 0 0 1 14.25 15H1.75A1.75 1.75 0 0 1 0 13.25v-2.5c0-.966.784-1.75 1.75-1.75Zm2.25 2.5a1 1 0 1 0 0 2 1 1 0 0 0 0-2Zm0-6.5a1 1 0 1 0 0-2 1 1 0 0 0 0 2Z"></path>',
    check:
      '<path d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L1.72 8.78a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z"></path>',
    alert:
      '<path d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575ZM9 11a1 1 0 1 0-2 0 1 1 0 0 0 2 0Zm-.25-5.25a.75.75 0 0 0-1.5 0v2.5a.75.75 0 0 0 1.5 0Z"></path>',
    x:
      '<path d="M2.343 13.657A8 8 0 1 1 13.658 2.343 8 8 0 0 1 2.343 13.657ZM6.03 4.97a.751.751 0 0 0-1.042.018.751.751 0 0 0-.018 1.042L6.94 8 4.97 9.97a.749.749 0 0 0 .326 1.275.749.749 0 0 0 .734-.215L8 9.06l1.97 1.97a.749.749 0 0 0 1.275-.326.749.749 0 0 0-.215-.734L9.06 8l1.97-1.97a.749.749 0 0 0-.326-1.275.749.749 0 0 0-.734.215L8 6.94Z"></path>',
    search:
      '<path d="M10.68 11.74a6 6 0 0 1-7.922-8.982 6 6 0 0 1 8.982 7.922l3.04 3.04a.749.749 0 0 1-.326 1.275.749.749 0 0 1-.734-.215ZM11.5 7a4.499 4.499 0 1 0-8.997 0A4.499 4.499 0 0 0 11.5 7Z"></path>',
    sync:
      '<path d="M1.705 8.005a.75.75 0 0 1 .834.656 5.5 5.5 0 0 0 9.592 2.97l-1.204-1.204a.25.25 0 0 1 .177-.427h3.646a.25.25 0 0 1 .25.25v3.646a.25.25 0 0 1-.427.177l-1.38-1.38A7.002 7.002 0 0 1 1.05 8.84a.75.75 0 0 1 .656-.834ZM8 2.5a5.487 5.487 0 0 0-4.131 1.869l1.204 1.204A.25.25 0 0 1 4.896 6H1.25A.25.25 0 0 1 1 5.75V2.104a.25.25 0 0 1 .427-.177l1.38 1.38A7.002 7.002 0 0 1 14.95 7.16a.75.75 0 0 1-1.49.178A5.5 5.5 0 0 0 8 2.5Z"></path>',
    chevron:
      '<path d="M6.22 3.22a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.75.75 0 0 1-1.06-1.06L9.94 8 6.22 4.28a.75.75 0 0 1 0-1.06Z"></path>',
    sun:
      '<path d="M8 12a4 4 0 1 1 0-8 4 4 0 0 1 0 8Zm0-1.5a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Zm5.657-8.157a.75.75 0 0 1 0 1.061l-1.061 1.06a.749.749 0 0 1-1.275-.326.749.749 0 0 1 .215-.734l1.06-1.06a.75.75 0 0 1 1.06 0Zm-9.193 9.193a.75.75 0 0 1 0 1.06l-1.06 1.061a.75.75 0 1 1-1.061-1.06l1.06-1.061a.75.75 0 0 1 1.061 0ZM8 0a.75.75 0 0 1 .75.75v1.5a.75.75 0 0 1-1.5 0V.75A.75.75 0 0 1 8 0ZM8 13a.75.75 0 0 1 .75.75v1.5a.75.75 0 0 1-1.5 0v-1.5A.75.75 0 0 1 8 13Zm8-5a.75.75 0 0 1-.75.75h-1.5a.75.75 0 0 1 0-1.5h1.5A.75.75 0 0 1 16 8ZM3 8a.75.75 0 0 1-.75.75H.75a.75.75 0 0 1 0-1.5h1.5A.75.75 0 0 1 3 8Zm10.657 5.657a.75.75 0 0 1-1.061 0l-1.06-1.061a.75.75 0 0 1 1.06-1.06l1.06 1.06a.75.75 0 0 1 0 1.061ZM4.464 4.464a.75.75 0 0 1-1.06 0L2.343 3.404a.75.75 0 0 1 1.06-1.06l1.061 1.06a.75.75 0 0 1 0 1.06Z"></path>',
    moon:
      '<path d="M9.598 1.591a.749.749 0 0 1 .785-.175 7.001 7.001 0 1 1-8.967 8.967.75.75 0 0 1 .961-.96 5.5 5.5 0 0 0 7.046-7.046.75.75 0 0 1 .175-.786Z"></path>',
    clock:
      '<path d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0ZM1.5 8a6.5 6.5 0 1 0 13 0 6.5 6.5 0 0 0-13 0Zm7-3.25v2.992l2.028.812a.75.75 0 0 1-.557 1.392l-2.5-1A.751.751 0 0 1 7 8.25v-3.5a.75.75 0 0 1 1.5 0Z"></path>'
  };

  function icon(name, size) {
    var body = icons[name] || icons.box;
    var s = size || 16;
    return (
      '<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="' +
      s +
      '" height="' +
      s +
      '" fill="currentColor" aria-hidden="true">' +
      body +
      "</svg>"
    );
  }

  var navGroups = [
    {
      heading: "Workspace",
      items: [
        { id: "policies", label: "Policy", icon: "shield", href: "policies.html" },
        { id: "fleet", label: "Fleet", icon: "fleet", href: "dashboard.html" },
        { id: "alerts", label: "Alerts", icon: "bell", href: "alerts.html", counter: "3", tone: "counter-attention" },
        { id: "agents", label: "Agents", icon: "people", href: "agents.html", counter: "2" }
      ]
    },
    {
      heading: "Administration",
      items: [
        { id: "users", label: "Users", icon: "person", href: "users.html" },
        { id: "audit", label: "Audit log", icon: "log", href: "audit.html" }
      ]
    }
  ];

  function sidebar(active) {
    var html =
      '<aside class="sidebar">' +
      '<a class="brand" href="index.html">' +
      '<span class="brand-mark">' +
      icon("box") +
      "</span>" +
      '<span><span class="brand-name">kfleet</span><br><span class="brand-sub">Control plane</span></span>' +
      "</a>";

    navGroups.forEach(function (group) {
      html += '<p class="nav-heading">' + group.heading + "</p><ul class=\"nav\">";
      group.items.forEach(function (item) {
        var current = item.id === active ? ' aria-current="page"' : "";
        html +=
          '<li><a class="nav-item" href="' +
          item.href +
          '"' +
          current +
          ">" +
          icon(item.icon) +
          "<span>" +
          item.label +
          "</span>" +
          (item.counter
            ? '<span class="counter ' + (item.tone || "") + '">' + item.counter + "</span>"
            : "") +
          "</a></li>";
      });
      html += "</ul>";
    });

    html +=
      '<div class="sidebar-footer">' +
      '<div class="user-card">' +
      '<span class="avatar">SW</span>' +
      '<span><span style="font-weight:600">solomon</span><br><span class="muted" style="font-size:12px">Admin</span></span>' +
      "</div>" +
      '<div class="user-card" style="justify-content:space-between">' +
      '<span class="muted" style="font-size:12px">Hub status</span>' +
      '<span class="row" style="gap:6px"><span class="dot dot-success"></span><span style="font-size:12px">Connected</span></span>' +
      "</div>" +
      "</div></aside>";

    return html;
  }

  function topbar(crumbs, actions) {
    var parts = (crumbs || "kfleet / Fleet").split("/");
    var trail = parts
      .map(function (part, index) {
        var text = part.trim();
        return index === parts.length - 1
          ? '<span style="color:var(--fgColor-default);font-weight:600">' + text + "</span>"
          : "<span>" + text + "</span>" + icon("chevron", 12);
      })
      .join("");

    return (
      '<header class="topbar">' +
      '<nav class="breadcrumb" aria-label="Breadcrumb">' +
      trail +
      "</nav>" +
      '<div class="topbar-actions">' +
      (actions || "") +
      '<button class="btn btn-icon" type="button" data-theme-toggle aria-label="Toggle color mode">' +
      '<span data-theme-icon>' +
      icon("moon") +
      "</span></button>" +
      "</div></header>"
    );
  }

  function applyTheme(mode) {
    var root = document.documentElement;
    root.setAttribute("data-color-mode", mode);
    root.setAttribute("data-light-theme", "light");
    root.setAttribute("data-dark-theme", "dark");
    try {
      localStorage.setItem("kfleet-mockup-theme", mode);
    } catch (error) {
      /* preview may run without storage access */
    }
    var slot = document.querySelector("[data-theme-icon]");
    if (slot) slot.innerHTML = icon(mode === "dark" ? "sun" : "moon");
  }

  function storedTheme() {
    try {
      return localStorage.getItem("kfleet-mockup-theme") || "light";
    } catch (error) {
      return "light";
    }
  }

  function mount() {
    var body = document.body;
    var shell = body.getAttribute("data-shell");

    if (shell !== "none") {
      var host = document.querySelector("[data-app]");
      if (host) {
        host.insertAdjacentHTML("afterbegin", sidebar(body.getAttribute("data-page")));
        var main = host.querySelector(".main");
        if (main) {
          main.insertAdjacentHTML(
            "afterbegin",
            topbar(body.getAttribute("data-breadcrumb"), body.getAttribute("data-actions"))
          );
        }
      }
    } else {
      var floating = document.querySelector("[data-theme-slot]");
      if (floating) {
        floating.innerHTML =
          '<button class="btn btn-icon" type="button" data-theme-toggle aria-label="Toggle color mode"><span data-theme-icon></span></button>';
      }
    }

    document.querySelectorAll("[data-icon]").forEach(function (node) {
      node.innerHTML = icon(node.getAttribute("data-icon"), Number(node.getAttribute("data-icon-size")) || 16);
    });

    applyTheme(storedTheme());

    document.addEventListener("click", function (event) {
      var toggle = event.target.closest("[data-theme-toggle]");
      if (!toggle) return;
      applyTheme(document.documentElement.getAttribute("data-color-mode") === "dark" ? "light" : "dark");
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", mount);
  } else {
    mount();
  }
})();
