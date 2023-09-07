// thank you https://stackoverflow.com/a/15289883
function dateDiffInDays(a, b) {
  const _MS_PER_DAY = 1000 * 60 * 60 * 24;
  // Discard the time and time-zone information.
  const utc1 = Date.UTC(a.getFullYear(), a.getMonth(), a.getDate());
  const utc2 = Date.UTC(b.getFullYear(), b.getMonth(), b.getDate());

  return Math.floor((utc2 - utc1) / _MS_PER_DAY);
}

// very rough results are totally acceptable here, a day or two is not
// important, considering timezones
function dateDiffString(now, other) {
  let diff = dateDiffInDays(now, other);
  if (diff < 0) {
    return '';
  }

  switch (diff) {
    case 0:
      return ' (today!)'
    case 1:
      return ' (tomorrow)'
    default:
      if (diff > 60) {
        let months = Math.floor(diff / 30.5);
        return ` (in ${months} months)`;
      }

      return ` (in ${diff} days)`
  }
}

function showReleaseInfoPopover(link) {
  let cell = link.closest('th');
  let release = cell.dataset.release;
  let template = document.querySelector('.release-popover-template').cloneNode(true);
  let now = new Date();

  let releaseDate = 'TBD';
  if (cell.dataset.releaseDate) {
    let parsed = new Date(cell.dataset.releaseDate);
    releaseDate = parsed.toDateString() + dateDiffString(now, parsed);
  }

  let eolDate = 'TBD';
  if (cell.dataset.eolDate) {
    let parsed = new Date(cell.dataset.eolDate);
    eolDate = parsed.toDateString() + dateDiffString(now, parsed);
  }

  let latestVersion = '(unreleased)';
  if (cell.dataset.latestVersion) {
    latestVersion = cell.dataset.latestVersion;
  }

  // Hide additional infos if the release isn't out yet;
  // do not rely on the browser date, as some release information
  // depends on buildtime data and so making it client-dependent
  // would not magically make more data appear.
  let container = template.querySelector('.release-links');
  container.classList.toggle('unreleased', cell.dataset.released !== 'true');

  template.querySelector('.release-date').innerText = releaseDate;
  template.querySelector('.latest-version').innerText = latestVersion;
  template.querySelector('.eol-date').innerText = eolDate;

  template.querySelector('.release-documentation').href = `apidocs/${release}/`;
  template.querySelector('.release-changelog').href = `https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-${release}.md`;
  template.querySelector('.release-gitbranch').href = `https://github.com/kubernetes/kubernetes/tree/release-${release}`;

  return template.innerHTML;
}

var visiblePopover = null;

const popoverTriggerList = document.querySelectorAll('thead th.release [data-bs-toggle="popover"]')
const popoverList = [...popoverTriggerList].map(function (popoverTriggerEl) {
  let popper = new bootstrap.Popover(popoverTriggerEl, {
    html:      true,
    content:   showReleaseInfoPopover,
    placement: 'bottom',
    trigger:   'manual',
  });

  popoverTriggerEl.addEventListener('click', function(e) {
    if (visiblePopover !== null) {
      visiblePopover.hide();
    }

    popper.show();
    visiblePopover = popper;

    // hijack the click event
    e.preventDefault();

    // prevent the body's listener from immediately closing the popover again
    e.stopPropagation();
  });
});

// close popovers _only_ when clicking outside them
document.body.addEventListener('click', function(e) {
  let target = e.target;

  // no popover visible
  if (visiblePopover === null) {
    return;
  }

  // clicked on a popover trigger element, ignore
  if (target.dataset.toggle === 'popover') {
    return;
  }

  // clicked inside a popover
  if (target.closest('.popover') !== null) {
    return;
  }

  visiblePopover.hide();
  visiblePopover = null;
});

// tbodies cannot be nested, so besides 1 tbody per APIGroup, there is no
// further nesting in the megatable; but APIResources should still logically
// be grouped below APIVersions, so to simulate this, we have to manually
// apply the "hidden" class for all APIResources which have their APIVersion
// collapsed;
// This "hidden" class has then also be kept in-sync with the collapsed status.
function updateAPIResourcesVisibility(apiversionRow) {
  let isCollapsed     = apiversionRow.classList.contains("collapsed");
  let apiVersion      = apiversionRow.dataset.apiversion;
  let apiGroupBody    = apiversionRow.closest('tbody');
  let apiResourceRows = apiGroupBody.querySelectorAll('tr.apiresource[data-apiversion="' + apiVersion + '"]');

  apiResourceRows.forEach(function(row) {
    row.classList.toggle("hidden", isCollapsed);
  });
}

// initial setup
document.querySelectorAll('tr.apiversion').forEach(updateAPIResourcesVisibility);

// toggle visibility of APIVersions for an APIGroup
document.querySelectorAll('tr.apigroup a.toggle').forEach(function(node) {
  node.addEventListener("click", function(e) {
    let apigroupRow  = this.closest('tr');
    let apigroupBody = apigroupRow.closest('tbody');
    let apigroup     = apigroupBody.dataset.apigroup;

    // toggle visibility in legend table
    apigroupBody.classList.toggle("collapsed");

    // update icon
    this.innerText = apigroupBody.classList.contains("collapsed") ? '⊕' : '⊙';

    e.preventDefault();
  });
});

// toggle visibility of APIResources for an APIVersion
document.querySelectorAll('tr.apiversion a.toggle').forEach(function(node) {
  node.addEventListener("click", function(e) {
    let apiversionRow = this.closest('tr');
    let apiversion    = apiversionRow.dataset.apiversion;
    let apigroupBody  = apiversionRow.closest('tbody');
    let apigroup      = apigroupBody.dataset.apigroup;

    // toggle visibility in legend table
    apiversionRow.classList.toggle("collapsed");
    updateAPIResourcesVisibility(apiversionRow);

    // update icon
    this.innerText = apiversionRow.classList.contains("collapsed") ? '⊕' : '⊙';

    e.preventDefault();
  });
});

// handle ROI dropdown changes
let selector = document.querySelector('#roi-selector');
let megatable = document.querySelector('#release-megatable');
let releaseColumns = document.querySelectorAll('th.release');

function roiClass(release) {
  return 'roi-' + release.replace('.', '-');
}

function updateROIState() {
  let selectedRelease = selector.value;
  let isSelected = selectedRelease != '';

  megatable.classList.toggle('roi-mode', isSelected);
  megatable.classList.toggle('non-roi-mode', !isSelected);
  megatable.classList.toggle('container-xxl', isSelected);

  // disable the archive view switch for UI clarity
  if (isSelected) {
    archiveViewSwitch.setAttribute('disabled', 'disabled');
  } else {
    archiveViewSwitch.removeAttribute('disabled');
  }

  // reset any other class
  releaseColumns.forEach(function(col) {
    megatable.classList.remove('show-' + roiClass(col.dataset.release));
  });

  // set the currently selected column
  let selectedRoiClass = roiClass(selectedRelease);
  if (isSelected) {
    megatable.classList.add('show-' + selectedRoiClass);
  }

  // if after all this no rows are visible, show a special row that
  // mentions no breaks
  let unremarkable = megatable.querySelectorAll('tbody tr.apigroup.' + selectedRoiClass).length === 0;
  megatable.classList.toggle('unremarkable', isSelected && unremarkable);

  // re-set columns from the archive view mode
  updateArchiveViewState();
}

selector.addEventListener('change', updateROIState);

// handle archiveView switch being toggled
let archiveViewSwitch = document.querySelector('#archiveViewSwitch');

function updateArchiveViewState() {
  megatable.classList.toggle('show-archive', archiveViewSwitch.checked);
  megatable.classList.toggle('hide-archive', !archiveViewSwitch.checked);

  if (megatable.classList.contains('non-roi-mode')) {
    megatable.classList.toggle('container-xxl', !archiveViewSwitch.checked);
  }
}

archiveViewSwitch.addEventListener('change', updateArchiveViewState);

// initial setup
updateROIState();
updateArchiveViewState();
