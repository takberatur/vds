package com.agcforge.videodownloader.ui.fragment

import android.Manifest
import android.app.DownloadManager
import android.content.Context
import android.content.pm.PackageManager
import android.annotation.SuppressLint
import android.content.ClipboardManager
import android.net.Uri
import android.os.Bundle
import android.os.Environment
import android.text.Editable
import android.text.TextWatcher
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.GridLayoutManager
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.data.model.Platform
import com.agcforge.videodownloader.databinding.FragmentHomeBinding
import com.agcforge.videodownloader.ui.adapter.PlatformAdapter
import com.agcforge.videodownloader.ui.viewmodel.HomeViewModel
import com.agcforge.videodownloader.utils.AppManager
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.Resource
import com.agcforge.videodownloader.utils.showToast
import com.google.android.material.dialog.MaterialAlertDialogBuilder
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

class HomeFragment : Fragment() {

    private var _binding: FragmentHomeBinding? = null
    private val binding get() = _binding!!

    private val viewModel: HomeViewModel by viewModels()
    private lateinit var platformAdapter: PlatformAdapter
    private var selectedPlatform: Platform? = null
	private var isSubmitting = false
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
        _binding = FragmentHomeBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
		preferenceManager = PreferenceManager(requireContext())

        setupRecyclerView()
        setupListeners()
        observeViewModel()

        viewModel.loadPlatforms()
		updateDownloadButtonState()
    }

    @SuppressLint("SetTextI18n")
    private fun setupRecyclerView() {
        platformAdapter = PlatformAdapter { platform ->
            // Handle platform click - auto-fill platform info
            selectedPlatform = platform
            platformAdapter.setSelection(platform)
            binding.tvPlatformName.text = "${getString(R.string.download_video)} - ${platform.name}"
            requireContext().showToast("Selected: ${platform.name}")
			updateDownloadButtonState()
        }

        binding.rvPlatforms.apply {
            adapter = platformAdapter
            layoutManager = GridLayoutManager(requireContext(), 2)
        }
    }

    private fun setupListeners() {
        binding.btnDownload.setOnClickListener {
            val url = binding.etUrl.text.toString().trim()
			if (!isUrlValid(url)) {
				binding.tilUrl.error = "Masukkan URL yang valid"
				updateDownloadButtonState()
				return@setOnClickListener
			}

            val platform = selectedPlatform
            if (platform == null) {
                requireContext().showToast("Please select a platform")
                return@setOnClickListener
            }

			binding.tilUrl.error = null
            viewModel.createDownload(url, platform.type)
        }

		binding.etUrl.addTextChangedListener(object : TextWatcher {
			override fun beforeTextChanged(s: CharSequence?, start: Int, count: Int, after: Int) = Unit
			override fun onTextChanged(s: CharSequence?, start: Int, before: Int, count: Int) = Unit
			override fun afterTextChanged(s: Editable?) {
				if (binding.tilUrl.error != null) binding.tilUrl.error = null
				updateDownloadButtonState()
			}
		})

        binding.etUrl.setOnFocusChangeListener { _, hasFocus ->
            if (!hasFocus) {
                binding.tilUrl.error = null
            }
        }

        binding.btnPaste.setOnClickListener {
            val clipboard = requireContext().getSystemService(ClipboardManager::class.java)
            val clip = clipboard.primaryClip
            if (clip != null && clip.itemCount > 0) {
                val text = clip.getItemAt(0).text?.toString()
                if (!text.isNullOrEmpty()) {
                    binding.etUrl.setText(text)
                    requireContext().showToast("Pasted: $text", Toast.LENGTH_SHORT)
					updateDownloadButtonState()
                }
            }
        }
    }

    private fun observeViewModel() {
        viewLifecycleOwner.lifecycleScope.launch {
            viewModel.platforms.collect { resource ->
                when (resource) {
                    is Resource.Loading -> {
                        binding.progressBar.visibility = View.VISIBLE
                    }
                    is Resource.Success -> {
                        binding.progressBar.visibility = View.GONE
                        resource.data?.let { platformAdapter.submitList(it) }
                    }
                    is Resource.Error -> {
                        binding.progressBar.visibility = View.GONE
                        requireContext().showToast(resource.message ?: "Failed to load platforms")
                    }

                    else -> {}
                }
            }
        }

        viewLifecycleOwner.lifecycleScope.launch {
            viewModel.downloadResult.collect { resource ->
                when (resource) {
                    is Resource.Loading -> {
						isSubmitting = true
                        binding.btnDownload.isEnabled = false
                        binding.progressBar.visibility = View.VISIBLE
                    }
                    is Resource.Success -> {
						isSubmitting = false
                        binding.btnDownload.isEnabled = true
                        binding.progressBar.visibility = View.GONE

                        resource.data?.let { task ->
                            // Show format selection dialog if formats available
                            if (!task.formats.isNullOrEmpty()) {
                                showFormatSelectionDialog(task)
                            } else {
                                requireContext().showToast("Download started!")
                                binding.etUrl.text?.clear()
							updateDownloadButtonState()
                            }
                        }
                    }
                    is Resource.Error -> {
						isSubmitting = false
                        binding.btnDownload.isEnabled = true
                        binding.progressBar.visibility = View.GONE
                        requireContext().showToast(resource.message ?: "Download failed")
						updateDownloadButtonState()
                    }
					is Resource.Idle -> {
						isSubmitting = false
						binding.progressBar.visibility = View.GONE
						updateDownloadButtonState()
					}
                }
            }
        }
    }

	private fun isUrlValid(url: String): Boolean {
		if (url.isBlank()) return false
		if (!url.startsWith("http://") && !url.startsWith("https://")) return false
		return runCatching { android.net.Uri.parse(url) }.isSuccess
	}

	private fun updateDownloadButtonState() {
		if (isSubmitting) {
			binding.btnDownload.isEnabled = false
			return
		}

		val url = binding.etUrl.text?.toString()?.trim().orEmpty()
		val enabled = url.isNotBlank()
		binding.btnDownload.isEnabled = enabled
	}

    private fun showFormatSelectionDialog(task: com.agcforge.videodownloader.data.model.DownloadTask) {
        val formats = task.formats ?: return
        val formatNames = formats.map { it.getFormatDescription() }.toTypedArray()

        MaterialAlertDialogBuilder(requireContext())
            .setTitle("Select Quality")
            .setItems(formatNames) { dialog, which ->
                val selectedFormat = formats[which]
				enqueueDownload(buildProxyVideoUrl(task, selectedFormat))
				requireContext().showToast("Download dimulai")
                binding.etUrl.text?.clear()
				updateDownloadButtonState()
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
			} catch (e: Exception) {
				requireContext().showToast(e.message ?: "Gagal memulai download")
			}
		}
	}

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
