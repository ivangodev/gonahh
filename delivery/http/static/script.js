categoryClass = "Category rounded-md border-2 w-fit border-sky-600 p-0.5 bg-sky-400 hover:bg-sky-600 mr-4 mt-4 font-serif italic"
topClass = "Top mt-10"
categoriesClass = "w-full mx-auto flex flex-wrap"

function getTopIdByCateg(categName) {
  return "top" + categName.replace(/ /g, '')
}

function cleanPrevReq() {
  searchBox = document.getElementById("SearchBox")
  for (s = searchBox.nextSibling; s != null; s = searchBox.nextSibling) {
    s.remove()
  }
}

function renderNotFound() {
  html = `<div class="w-full mx-auto mt-8 font-mono text-center">Nothing was found. Try another request.</div>`
  s = document.getElementById("SearchBox");
  s.insertAdjacentHTML('afterend', html)
}

function doReq(jobname) {
  var xmlHttp = new XMLHttpRequest();
  url = document.URL + '/api?jobname=' + encodeURIComponent(jobname);
  xmlHttp.open("GET", url, false);
  xmlHttp.send(null);
  return JSON.parse(xmlHttp.responseText)
}

function findClicked() {
  jobname = document.getElementById("searchBoxContent").value
  resp = doReq(jobname)
  resp.CategoriesTop.unshift({
    Category: "All",
    Top: resp.AllTop
  })
  cleanPrevReq()
  if (resp.JobsNumber != 0)
    renderCategsAndTops(resp.CategoriesTop)
  else
    renderNotFound()
}

function selectCategory(clickedCateg) {
  categs = document.getElementsByClassName("Category")
  for (let i = 0; i < categs.length; i++) {
    if (categs.item(i) != clickedCateg) {
      categs.item(i).setAttribute("class", categoryClass)
    } else {
      clickedCateg.setAttribute("class", categoryClass + " underline")
    }
  }

  tops = document.getElementsByClassName("Top")
  id = getTopIdByCateg(clickedCateg.textContent)
  selectedTop = document.getElementById(id)
  for (let i = 0; i < tops.length; i++) {
    if (tops.item(i) != selectedTop) {
      tops.item(i).setAttribute("class", topClass + " hidden")
    } else {
      selectedTop.setAttribute("class", topClass)
    }
  }

}

function selectCategoryAll() {
  categs = document.getElementsByClassName("Category")
  //All MUST go first
  selectCategory(categs.item(0))
}

function renderCategories(categNames) {
  categs = ""
  for (let i = 0; i < categNames.length; i++) {
    categs += `<div onclick="selectCategory(this)" class="${categoryClass}">${categNames[i]}</div>`
  }
  res = `<div class="${categoriesClass}" id="Categories">${categs}</div>`

  searchBox = document.getElementById("SearchBox")
  searchBox.insertAdjacentHTML('afterend', res)
}

function aTopPosHTML(categName, topPosObj, num, maxRate) {
  relRate = Math.floor((topPosObj.Rate / maxRate) * 100)
  revRate = 100 - relRate
  color = ""
  if (num % 2 == 0) {
    color = "bg-gray-100"
  } else {
    color = "bg-white"
  }

  res = `
<div class="w-full mx-auto flex mt-4 ${color} border-b-2">
	<div class="w-7 font-mono">${num+1}.</div>
	<div class="w-36 mr-7 flex justify-between">
		<div class="font-mono">${topPosObj.Keyword}</div>
		<div class="font-mono">${topPosObj.Rate}%</div>
	</div>
	<div class="flex grow items-center  mr-1">
		<div class="basis-[${relRate}%] bg-blue-600 h-3">&nbsp</div>
		<div class="basis-[${revRate}%] bg-blue-400 h-3"></div>
	</div>
</div>`

  return res
}

function renderTop(categName, top) {
  html = ""
  for (let i = 0; i < top.length; i++) {
    html += aTopPosHTML(categName, top[i], i, top[0].Rate)
  }
  id = getTopIdByCateg(categName)
  res = `<div class="${topClass}" id=${id}>${html}</div>`

  categsDiv = document.getElementById("Categories");
  categsDiv.insertAdjacentHTML('afterend', res)
}

function renderCategsAndTops(tops) { //CategoriesTop
  categs = []
  for (let i = 0; i < tops.length; i++) {
    categs.push(tops[i].Category)
  }
  renderCategories(categs)

  for (let i = 0; i < tops.length; i++) {
    renderTop(tops[i].Category, tops[i].Top)
  }

  selectCategoryAll()
}
