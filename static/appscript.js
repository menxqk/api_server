const logout = document.getElementById('logout');
const logoutModal = document.getElementById('logout-modal');
const loaderModal = document.getElementById('loader-modal');
// const btnYes = document.getElementById('btn-yes');
const btnNo = document.getElementById('btn-no');
const bucketsBtn = document.getElementById('buckets');
const collectionsBtn = document.getElementById('collections');
const mainContent = document.querySelector('main .content');

logout.addEventListener('click', (e) => {
    e.preventDefault();
    logoutModal.classList.add('show');
});

btnNo.addEventListener('click', () => {
    logoutModal.classList.remove('show');
});

// btnYes.addEventListener('click', () => {
//     window.location.href = '/logout';
// });

bucketsBtn.addEventListener('click', (e) => {
    e.preventDefault();
    showLoader();
    showBuckets();
});

collectionsBtn.addEventListener('click', (e) => {
    e.preventDefault();
    showLoader();
    showCollections();
});

function showLoader() {
    loaderModal.classList.add('show');
}

function removeLoader() {
    loaderModal.classList.remove('show');
}

async function showBuckets() {
    mainContent.innerHTML = `<h3>Buckets</h3><br/>`;
    const list = document.createElement('ul');
    mainContent.appendChild(list);
    const resp = await fetch('http://localhost:8080/app/api/buckets/');
    const buckets = await resp.json();

    if (Array.isArray(buckets)) {
        buckets.forEach((bucket, index, arr) => {
            const listItem = document.createElement('li');
            const a = document.createElement('a');
            a.href = `/app/api/buckets/${bucket}`;
            a.onclick = function(e) { 
                e.preventDefault(); 
                showLoader(); 
                showBucket(`${bucket}`);
            };
            a.textContent = `${bucket}/`;
            listItem.appendChild(a);
            list.appendChild(listItem);
        });
    }
    removeLoader();
}

async function showBucket(bucketName) {
    // mainContent.innerHTML = `
    // <h3>Bucket ${bucketName}</h3>
    // <br/>
    // <form id="upload-form" enctype="multipart/form-data" action="/app/api/buckets/${bucketName}" method="post">
    //     <input id="upload-file" type="file" onchange="enableUpload()" name="upload-file" />
    //     <input id="upload-name" type="hidden" value="" name="upload-name" />
    //     <input id="upload-size" type="hidden" value="" name="upload-size" />
    //     <input id="upload-type" type="hidden" value="" name="upload-type" />   
    //     <input id="upload-submit" onclick="doFileUpload(e, ${bucketName})" type="submit" value="Upload File" disabled/>
    // </form>
    // <br/>
    // `;
    mainContent.innerHTML = `
    <h3>Bucket ${bucketName}</h3>
    <br/>
    <form id="upload-form" method="post" enctype="multipart/form-data">
        <input id="upload-file" type="file" onchange="enableUpload()" name="upload-file" />
        <input id="upload-name" type="hidden" value="" name="upload-name" />
        <input id="upload-size" type="hidden" value="" name="upload-size" />
        <input id="upload-type" type="hidden" value="" name="upload-type" />   
        <input id="upload-submit" type="submit" value="Upload File" disabled/>
    </form>
    <br/>
    `;

    const form = document.getElementById('upload-form');
    form.addEventListener('submit', (e) => {
        e.preventDefault();
        doFileUpload(bucketName);
    });
    const list = document.createElement('ul');
    mainContent.appendChild(list)
    const resp = await fetch(`http://localhost:8080/app/api/buckets/${bucketName}`);
    const objects = await resp.json();

    if (Array.isArray(objects)) {
        objects.forEach((object, index, arr) => {
            const listItem = document.createElement('li');
            const a = document.createElement('a');
            a.href = `/app/api/buckets/${bucketName}/${object}`;
            a.textContent = `${object}`;
            const space = document.createElement('span');
            space.innerText = " ";
            const x = document.createElement('a');
            x.href = '#';
            x.textContent = "Apagar";
            x.onclick = function(e) {
                e.preventDefault();
                showLoader();
                deleteObject(`${bucketName}`, `${object}`);
            }
            listItem.appendChild(a);
            listItem.appendChild(space);
            listItem.appendChild(x);
            list.appendChild(listItem);
        });
    }
    removeLoader();
}

async function deleteObject(bucketName, object) {
    config = {
        method: "DELETE"
    }
    const resp = await fetch(`http://localhost:8080/app/api/buckets/${bucketName}/${object}`, config)
    // console.log(resp.status, resp.statusText)
    removeLoader();
    window.location.replace(`http://localhost:8080/app`);
}


async function showCollections() {
    mainContent.innerHTML = `<h3>Collections</h3><br/>`;
    const list = document.createElement('ul');
    mainContent.appendChild(list);
    const resp = await fetch('http://localhost:8080/app/api/collections/');
    const collections = await resp.json();

    if (Array.isArray(collections)) {
        collections.forEach((collection, index, arr) => {
            const listItem = document.createElement('li');
            const a = document.createElement('a');
            a.href = `/app/api/collections/${collection}`;
            a.onclick = function(e) { e.preventDefault(); showCollection(`${collection}`);};
            a.textContent = `${collection}`;
            listItem.appendChild(a);            
            list.appendChild(listItem);
        });
    }
    removeLoader();
}

async function showCollection(collectionName) {
    mainContent.innerHTML = `<h3>Collection ${collectionName}</h3><br/>`;
    const list = document.createElement('ul');
    mainContent.appendChild(list)
    const resp = await fetch(`http://localhost:8080/app/api/collections/${collectionName}`);
    const docs = await resp.json();

    if (Array.isArray(docs)) {
        docs.forEach((doc, index, arr) => {
            const listItem = document.createElement('li');
            const a = document.createElement('a');
            a.href = `/app/api/collections/${collectionName}/${doc}`;
            a.textContent = `${doc}`;
            const space = document.createElement('span');
            space.innerText = " ";
            const x = document.createElement('a');
            x.href = '#';
            x.textContent = "Apagar";
            x.onclick = function(e) {
                e.preventDefault();
                showLoader();
                deleteDocument(`${collectionName}`, `${doc}`);
            }
            listItem.appendChild(a);
            listItem.appendChild(space);
            listItem.appendChild(x);
            list.appendChild(listItem);
        });
    }
    removeLoader();
}

async function deleteDocument(collectionName, doc) {
    config = {
        method: "DELETE"
    }
    const resp = await fetch(`http://localhost:8080/app/api/collections/${collectionName}/${doc}`, config)
    // console.log(resp.status, resp.statusText)
    removeLoader();
    window.location.replace(`http://localhost:8080/app`);
}

function enableUpload() {
    document.getElementById('upload-submit').disabled = false;
    // document.getElementById('upload-name').value = document.getElementById('upload-file').files.item(0).name;
    // document.getElementById('upload-size').value = document.getElementById('upload-file').files.item(0).size;
    // document.getElementById('upload-type').value = document.getElementById('upload-file').files.item(0).type;

    // console.log(document.getElementById('upload-name').value)
    // console.log(document.getElementById('upload-size').value)
    // console.log(document.getElementById('upload-type').value)
}


async function doFileUpload(bucketName) {
    const file = document.getElementById('upload-file').files[0];
    const name = document.getElementById('upload-file').files.item(0).name;
    const size = document.getElementById('upload-file').files.item(0).size;
    const type = document.getElementById('upload-file').files.item(0).type;
    const formData = new FormData();
    formData.append("upload-file", file);
    formData.append("upload-name", name);
    formData.append("upload-size", size);
    formData.append("upload-type", type);

    console.log(formData);
    config = {
        method: "POST",
        body: formData
    }

    try {
        const resp = await fetch(`http://localhost:8080/app/api/buckets/${bucketName}`, config);
        console.log('HTTP response code:', resp.status);
        window.location.replace(`http://localhost:8080/app`);
    } catch {
        console.log('Houston we have a problem...')
    }
    
}