document.addEventListener("DOMContentLoaded", () => {
  // Mark which page the user is on in the nav bar
  // I'm doing this dynamically to be lazy
  // Picking it server-side is annoying with the current setup
  document.querySelectorAll("nav a").forEach((elem) => {
    if (elem.getAttribute("href") === location.pathname) {
      elem.classList.add("current");
    }
  });
});
