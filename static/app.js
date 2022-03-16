const Controller = {
  search: (ev) => {
    ev.preventDefault()
    const form = document.getElementById("form")
    const data = Object.fromEntries(new FormData(form))
    const response = fetch(`/search?q=${data.query}`).then((response) => {
      response.json().then((results) => {
        Controller.updateHtml(results)
        Controller.highlightResults(data.query)
      });
    });
  },

  updateHtml: (results) => {
    document.getElementById("results").innerHTML = ''
    results.forEach(function (result, i) {
      const node = document.createElement("pre")
      const rule = document.createElement("hr")
      const textNode = document.createTextNode(result)
      node.appendChild(textNode)
      document.getElementById("results").appendChild(node)
      document.getElementById("results").appendChild(rule)
    })
  },

  highlightResults: (query) => {
    const rawResults = document.querySelectorAll("pre")
    const matcher = new RegExp(query, "gi")
    rawResults.forEach( function(element) {
        element.innerHTML = element.innerHTML.replace(matcher, "<mark>$&</mark>")
      }
    )
  }
};

const form = document.getElementById("form")
form.addEventListener("submit", Controller.search)
