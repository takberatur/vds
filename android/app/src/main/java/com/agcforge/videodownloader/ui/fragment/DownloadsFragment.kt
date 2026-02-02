package com.agcforge.videodownloader.ui.fragment

import android.Manifest
import android.app.DownloadManager
import android.content.Context
import android.content.pm.PackageManager
import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.os.Environment
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.LinearLayoutManager
import com.agcforge.videodownloader.databinding.FragmentDownloadsBinding
import com.agcforge.videodownloader.ui.adapter.DownloadTaskAdapter
import com.agcforge.videodownloader.ui.viewmodel.DownloadsViewModel
import com.agcforge.videodownloader.utils.AppManager
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.Resource
import com.agcforge.videodownloader.utils.StorageFolderNavigator
import com.agcforge.videodownloader.utils.showToast
import com.google.android.material.dialog.MaterialAlertDialogBuilder
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

class DownloadsFragment : Fragment() {

    private var _binding: FragmentDownloadsBinding? = null
    private val binding get() = _binding!!

    private val viewModel: DownloadsViewModel by viewModels()
    private lateinit var downloadAdapter: DownloadTaskAdapter
	private lateinit var preferenceManager: PreferenceManager
	private var pendingDownloadUrl: String? = null

	private val storagePermissionLauncher = registerForActivityResult(
		ActivityResultContracts.RequestPermission()
	) { granted ->
		val url = pendingDownloadUrl
		pendingDownloadUrl = null
		if (granted && url != null) {
			enqueueDownload(url)
		} else if (!granted) {
			requireContext().showToast("Izin penyimpanan ditolak")
		}
	}

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentDownloadsBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
		preferenceManager = PreferenceManager(requireContext())

        setupRecyclerView()
        setupSwipeRefresh()
        observeViewModel()
		binding.btnOpenStorageFolder.setOnClickListener { openStorageFolder() }

        viewModel.loadDownloads()
    }

	private fun openStorageFolder() {
		viewLifecycleOwner.lifecycleScope.launch {
			val location = preferenceManager.storageLocation.first() ?: "app"
			StorageFolderNavigator.openStorageFolder(requireContext(), location)
		}
	}

    private fun setupRecyclerView() {
        downloadAdapter = DownloadTaskAdapter(
            onItemClick = { task ->
                showDownloadDetails(task)
            },
            onDownloadClick = { task ->
                handleDownloadClick(task)
            }
        )

        binding.rvDownloads.apply {
            adapter = downloadAdapter
            layoutManager = LinearLayoutManager(requireContext())
        }
    }

    private fun setupSwipeRefresh() {
        binding.swipeRefresh.setOnRefreshListener {
            viewModel.loadDownloads()
        }
    }

    private fun observeViewModel() {
        viewLifecycleOwner.lifecycleScope.launch {
            viewModel.downloads.collect { resource ->
                binding.swipeRefresh.isRefreshing = false

                when (resource) {
                    is Resource.Loading -> {
                        binding.progressBar.visibility = View.VISIBLE
                        binding.tvEmpty.visibility = View.GONE
                    }
                    is Resource.Success -> {
                        binding.progressBar.visibility = View.GONE

                        resource.data?.let { downloads ->
                            if (downloads.isEmpty()) {
                                binding.tvEmpty.visibility = View.VISIBLE
                                binding.rvDownloads.visibility = View.GONE
                            } else {
                                binding.tvEmpty.visibility = View.GONE
                                binding.rvDownloads.visibility = View.VISIBLE
                                downloadAdapter.submitList(downloads)
                            }
                        }
                    }
                    is Resource.Error -> {
                        binding.progressBar.visibility = View.GONE
                        binding.tvEmpty.visibility = View.VISIBLE
                        requireContext().showToast(resource.message ?: "Failed to load downloads")
                    }

                    else -> {}
                }
            }
        }
    }

    private fun showDownloadDetails(task: com.agcforge.videodownloader.data.model.DownloadTask) {
        val details = buildString {
            append("Title: ${task.title ?: "Unknown"}\n\n")
            append("Platform: ${task.platformType}\n")
            append("Status: ${task.status}\n")
            append("Duration: ${task.getFormattedDuration()}\n")
            append("Size: ${task.fileSize?.let { "${it / (1024 * 1024)} MB" } ?: "N/A"}\n")
            append("Date: ${task.createdAt}\n")

            if (task.errorMessage != null) {
                append("\nError: ${task.errorMessage}")
            }
        }

        MaterialAlertDialogBuilder(requireContext())
            .setTitle("Download Details")
            .setMessage(details)
            .setPositiveButton("OK", null)
            .show()
    }

    private fun handleDownloadClick(task: com.agcforge.videodownloader.data.model.DownloadTask) {
        when (task.status.lowercase()) {
            "completed" -> {
                // Open downloaded file or show download options
                task.filePath?.let { path ->
                    openFile(path)
                } ?: run {
                    // Show format selection if available
                    if (!task.formats.isNullOrEmpty()) {
                        showFormatSelection(task)
                    }
                }
            }
            "failed" -> {
                requireContext().showToast("Download failed. Please try again.")
            }
            "processing" -> {
                requireContext().showToast("Download is being processed...")
            }
            else -> {
                requireContext().showToast("Download is pending...")
            }
        }
    }

    private fun showFormatSelection(task: com.agcforge.videodownloader.data.model.DownloadTask) {
        val formats = task.formats ?: return
        val formatNames = formats.map { it.getFormatDescription() }.toTypedArray()

        MaterialAlertDialogBuilder(requireContext())
            .setTitle("Select Download Quality")
            .setItems(formatNames) { dialog, which ->
                val selectedFormat = formats[which]
				val url = buildProxyVideoUrl(task, selectedFormat)
				downloadFile(url)
                dialog.dismiss()
            }
            .setNegativeButton("Cancel", null)
            .show()
    }

	private fun buildProxyVideoUrl(
		task: com.agcforge.videodownloader.data.model.DownloadTask,
		format: com.agcforge.videodownloader.data.model.DownloadFormat
	): String {
		val base = AppManager.baseUrl
		val endpoint = if (base.endsWith("/")) {
			"${base}public-proxy/downloads/file/video"
		} else {
			"$base/public-proxy/downloads/file/video"
		}

		val formatId = format.formatId
		val resolution = format.height?.let { "${it}p" }
		val effectiveFormat = formatId ?: resolution

		val filename = task.title?.takeIf { it.isNotBlank() } ?: "download"
		val uriBuilder = Uri.parse(endpoint).buildUpon()
			.appendQueryParameter("task_id", task.id)
			.appendQueryParameter("filename", filename)
		if (!effectiveFormat.isNullOrBlank()) {
			uriBuilder.appendQueryParameter("format_id", effectiveFormat)
		}
		return uriBuilder.build().toString()
	}

    private fun downloadFile(url: String) {
		enqueueDownload(url)
    }

	private fun enqueueDownload(url: String) {
		viewLifecycleOwner.lifecycleScope.launch {
			val storageLocation = preferenceManager.storageLocation.first() ?: "app"
			val uri = Uri.parse(url)
			val fileName = (uri.lastPathSegment?.takeIf { it.isNotBlank() } ?: "video_${System.currentTimeMillis()}.mp4")
				.let { if (it.contains('.')) it else "$it.mp4" }

			if (storageLocation == "downloads" && android.os.Build.VERSION.SDK_INT < 29) {
				val granted = ContextCompat.checkSelfPermission(
					requireContext(),
					Manifest.permission.WRITE_EXTERNAL_STORAGE
				) == PackageManager.PERMISSION_GRANTED
				if (!granted) {
					pendingDownloadUrl = url
					storagePermissionLauncher.launch(Manifest.permission.WRITE_EXTERNAL_STORAGE)
					return@launch
				}
			}

			val dm = requireContext().getSystemService(Context.DOWNLOAD_SERVICE) as DownloadManager
			val request = DownloadManager.Request(uri)
				.setTitle(fileName)
				.setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED)
				.setAllowedOverMetered(true)
				.setAllowedOverRoaming(true)

			if (storageLocation == "downloads") {
				request.setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, fileName)
			} else {
				request.setDestinationInExternalFilesDir(requireContext(), Environment.DIRECTORY_DOWNLOADS, fileName)
			}

			try {
				dm.enqueue(request)
				requireContext().showToast("Download dimulai")
			} catch (e: Exception) {
				requireContext().showToast(e.message ?: "Gagal memulai download")
			}
		}
	}

    private fun openFile(path: String) {
        try {
            val intent = Intent(Intent.ACTION_VIEW)
            intent.setDataAndType(Uri.parse(path), "video/*")
            intent.flags = Intent.FLAG_ACTIVITY_NEW_TASK
            startActivity(intent)
        } catch (e: Exception) {
            requireContext().showToast("No app found to open this file")
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
