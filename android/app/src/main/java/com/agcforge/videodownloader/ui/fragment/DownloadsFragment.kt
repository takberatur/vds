package com.agcforge.videodownloader.ui.fragment

import android.Manifest
import android.annotation.SuppressLint
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.activity.result.contract.ActivityResultContracts
import androidx.annotation.OptIn
import androidx.core.content.ContextCompat
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import androidx.media3.common.util.UnstableApi
import androidx.recyclerview.widget.LinearLayoutManager
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.data.model.LocalDownloadItem
import com.agcforge.videodownloader.databinding.FragmentDownloadsBinding
import com.agcforge.videodownloader.ui.activities.player.AudioPlayerActivity
import com.agcforge.videodownloader.ui.activities.player.VideoPlayerActivity
import com.agcforge.videodownloader.ui.adapter.LocalDownloadAdapter
import com.agcforge.videodownloader.ui.component.PlayerSelectionDialog
import com.agcforge.videodownloader.utils.LocalDownloadsScanner
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.StorageFolderNavigator
import com.agcforge.videodownloader.utils.showToast
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

class DownloadsFragment : Fragment() {

    private var _binding: FragmentDownloadsBinding? = null
    private val binding get() = _binding!!

	private lateinit var preferenceManager: PreferenceManager
	private lateinit var adapter: LocalDownloadAdapter

	private val readPermissionsLauncher = registerForActivityResult(
		ActivityResultContracts.RequestMultiplePermissions()
	) { _ ->
		loadLocalDownloads()
	}

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentDownloadsBinding.inflate(inflater, container, false)
        return binding.root
    }

    @SuppressLint("SuspiciousIndentation")
    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
		preferenceManager = PreferenceManager(requireContext())

        setupRecyclerView()
        setupSwipeRefresh()
		binding.btnOpenStorageFolder.setOnClickListener { openStorageFolder() }

		loadLocalDownloads()
    }

	private fun openStorageFolder() {
		viewLifecycleOwner.lifecycleScope.launch {
			val location = preferenceManager.storageLocation.first() ?: "app"
			StorageFolderNavigator.openStorageFolder(requireContext(), location)
		}
	}

    private fun setupRecyclerView() {
		adapter = LocalDownloadAdapter { item ->
			openLocalFile(item)
		}

        binding.rvDownloads.apply {
			adapter = this@DownloadsFragment.adapter
            layoutManager = LinearLayoutManager(requireContext())
        }
    }

    private fun setupSwipeRefresh() {
        binding.swipeRefresh.setOnRefreshListener {
			loadLocalDownloads()
        }
    }

	private fun loadLocalDownloads() {
		viewLifecycleOwner.lifecycleScope.launch {
			binding.progressBar.visibility = View.VISIBLE
			binding.tvEmpty.visibility = View.GONE
			binding.rvDownloads.visibility = View.GONE

			val location = preferenceManager.storageLocation.first() ?: "app"
			if (location == "downloads" && !hasRequiredReadPermissions()) {
				binding.progressBar.visibility = View.GONE
				binding.swipeRefresh.isRefreshing = false
				binding.tvEmpty.visibility = View.VISIBLE
				requestReadPermissions()
				return@launch
			}

			val items = LocalDownloadsScanner.scan(requireContext(), location)
			binding.progressBar.visibility = View.GONE
			binding.swipeRefresh.isRefreshing = false
			if (items.isEmpty()) {
				binding.tvEmpty.visibility = View.VISIBLE
				binding.rvDownloads.visibility = View.GONE
			} else {
				binding.tvEmpty.visibility = View.GONE
				binding.rvDownloads.visibility = View.VISIBLE
				adapter.submitList(items)
			}
		}
	}

	private fun hasRequiredReadPermissions(): Boolean {
		return if (android.os.Build.VERSION.SDK_INT >= 33) {
			ContextCompat.checkSelfPermission(requireContext(), Manifest.permission.READ_MEDIA_VIDEO) == PackageManager.PERMISSION_GRANTED &&
				ContextCompat.checkSelfPermission(requireContext(), Manifest.permission.READ_MEDIA_AUDIO) == PackageManager.PERMISSION_GRANTED
		} else {
			ContextCompat.checkSelfPermission(requireContext(), Manifest.permission.READ_EXTERNAL_STORAGE) == PackageManager.PERMISSION_GRANTED
		}
	}

	private fun requestReadPermissions() {
		val perms = if (android.os.Build.VERSION.SDK_INT >= 33) {
			arrayOf(Manifest.permission.READ_MEDIA_VIDEO, Manifest.permission.READ_MEDIA_AUDIO)
		} else {
			arrayOf(Manifest.permission.READ_EXTERNAL_STORAGE)
		}
		readPermissionsLauncher.launch(perms)
	}

    @OptIn(UnstableApi::class)
    private fun openLocalFile(item: LocalDownloadItem) {
        val isVideo = item.mimeType?.startsWith("video/") == true
        val playerType = if (isVideo) {
            PlayerSelectionDialog.PlayerType.VIDEO
        } else {
            PlayerSelectionDialog.PlayerType.AUDIO
        }

        PlayerSelectionDialog.Builder(requireContext())
            .setType(playerType)
            .setOnPlayerSelected { player ->
                when (player) {
                    PlayerSelectionDialog.PLAYER_INTERNAL -> {
                        if (isVideo) {
                            VideoPlayerActivity.start(requireContext(), item.uri.toString(), item.displayName)
                        } else {
                            AudioPlayerActivity.start(requireContext(), item.uri.toString(), item.displayName)
                        }
                    }
                    PlayerSelectionDialog.PLAYER_EXTERNAL -> {
                        val intent = Intent(Intent.ACTION_VIEW)
                        intent.setDataAndType(item.uri, item.mimeType)
                        intent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                        runCatching { startActivity(Intent.createChooser(intent, requireContext().getString(
                            R.string.open_with))) }
                            .onFailure {
                                requireContext().showToast(it.message ?: requireContext().getString(R.string.error_unknown))
                            }
                    }
                }
            }
            .show()
	}

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
