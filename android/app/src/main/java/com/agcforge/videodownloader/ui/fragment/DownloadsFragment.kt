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
import androidx.appcompat.app.AlertDialog
import androidx.core.content.ContextCompat
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import androidx.media3.common.util.UnstableApi
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
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
import java.io.File

class DownloadsFragment : Fragment() {

    private var _binding: FragmentDownloadsBinding? = null
    private val binding get() = _binding!!

	private lateinit var preferenceManager: PreferenceManager
	private lateinit var adapter: LocalDownloadAdapter

    private var isLoading = false
    private var hasMoreItems = true
    private var currentOffset = 0
    private val pageSize = 20

    private val readPermissionsLauncher = registerForActivityResult(
		ActivityResultContracts.RequestMultiplePermissions()
	) { _ ->
        resetAndLoadDownloads()
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

        resetAndLoadDownloads()
    }

	private fun openStorageFolder() {
		viewLifecycleOwner.lifecycleScope.launch {
			val location = preferenceManager.storageLocation.first() ?: "app"
			StorageFolderNavigator.openStorageFolder(requireContext(), location)
		}
	}

    private fun setupRecyclerView() {
        adapter = LocalDownloadAdapter(
            context = requireContext(),
            onOpenClick = { item ->
                openLocalFile(item)
            },
            onDeleteClick = { item ->
                showDeleteConfirmation(item)
            }
        )

        binding.rvDownloads.apply {
			adapter = this@DownloadsFragment.adapter
            layoutManager = LinearLayoutManager(requireContext())

            addOnScrollListener(object : RecyclerView.OnScrollListener() {
                private var lastVisibleRange = Pair(-1, -1)

                override fun onScrolled(recyclerView: RecyclerView, dx: Int, dy: Int) {
                    super.onScrolled(recyclerView, dx, dy)

                    val layoutManager = recyclerView.layoutManager as LinearLayoutManager
                    val firstVisible = layoutManager.findFirstVisibleItemPosition()
                    val lastVisible = layoutManager.findLastVisibleItemPosition()

                    if (firstVisible != lastVisibleRange.first || lastVisible != lastVisibleRange.second) {
                        lastVisibleRange = Pair(firstVisible, lastVisible)
                        loadThumbnailsForVisibleRange(firstVisible, lastVisible)
                    }

                    if (!isLoading && hasMoreItems) {
                        val visibleItemCount = layoutManager.childCount
                        val totalItemCount = layoutManager.itemCount

                        if ((visibleItemCount + firstVisible) >= totalItemCount - 5) {
                            loadMoreDownloads()
                        }
                    }
                }

                override fun onScrollStateChanged(recyclerView: RecyclerView, newState: Int) {
                    super.onScrollStateChanged(recyclerView, newState)

                    if (newState == RecyclerView.SCROLL_STATE_IDLE) {
                        val layoutManager = recyclerView.layoutManager as LinearLayoutManager
                        val firstVisible = layoutManager.findFirstVisibleItemPosition()
                        val lastVisible = layoutManager.findLastVisibleItemPosition()
                        loadThumbnailsForVisibleRange(firstVisible, lastVisible)
                    }
                }
            })

            itemAnimator = null
        }
    }

    private fun setupSwipeRefresh() {
        binding.swipeRefresh.setOnRefreshListener {
            resetAndLoadDownloads()
        }
    }

    private fun resetAndLoadDownloads() {
        currentOffset = 0
        hasMoreItems = true
        adapter.clearItems()
        loadLocalDownloads()
    }

	@SuppressLint("SuspiciousIndentation")
    private fun loadLocalDownloads() {
        if (isLoading) return

        isLoading = true
        binding.progressBar.visibility = View.VISIBLE
        binding.tvEmpty.visibility = View.GONE

        println("DEBUG: Loading downloads - offset: $currentOffset, hasMore: $hasMoreItems")

        viewLifecycleOwner.lifecycleScope.launch {
            val location = preferenceManager.storageLocation.first() ?: "app"

            println("DEBUG: Storage location: $location")

            if (location == "downloads" && !hasRequiredReadPermissions()) {
                binding.progressBar.visibility = View.GONE
                binding.swipeRefresh.isRefreshing = false
                binding.tvEmpty.visibility = View.VISIBLE
                requestReadPermissions()
                isLoading = false
                return@launch
            }

            try {
                println("DEBUG: Calling scanPaged with limit=$pageSize, offset=$currentOffset")

                val result = LocalDownloadsScanner.scanPaged(
                    context = requireContext(),
                    location = location,
                    limit = pageSize,
                    offset = currentOffset
                )

                println("DEBUG: Got ${result.items.size} items, hasMore: ${result.hasMore}")

                if (result.items.isNotEmpty()) {
                    adapter.addItems(result.items)
                    binding.rvDownloads.visibility = View.VISIBLE
                    binding.tvEmpty.visibility = View.GONE

                    binding.rvDownloads.post {
                        val layoutManager = binding.rvDownloads.layoutManager as LinearLayoutManager
                        val firstVisible = layoutManager.findFirstVisibleItemPosition()
                        val lastVisible = layoutManager.findLastVisibleItemPosition()

                        if (firstVisible >= 0 && lastVisible >= firstVisible) {
                            loadThumbnailsForVisibleRange(firstVisible, lastVisible)
                        }
                    }

                    if (result.items.isNotEmpty()) {
                        println("DEBUG: First item: ${result.items[0].displayName}")
                    }
                } else if (currentOffset == 0) {
                    println("DEBUG: No items found at offset 0")
                    binding.tvEmpty.visibility = View.VISIBLE
                    binding.rvDownloads.visibility = View.GONE
                    binding.tvEmpty.text = getString(R.string.no_media_files_found)
                }

                hasMoreItems = result.hasMore
                currentOffset = result.nextOffset

                println("DEBUG: Updated - currentOffset: $currentOffset, hasMore: $hasMoreItems")

                if (result.items.isNotEmpty()) {
                    loadThumbnailsForVisibleItems()
                }

            } catch (e: Exception) {
                e.printStackTrace()
                println("DEBUG: Error loading downloads: ${e.message}")
                requireContext().showToast(getString(R.string.failed_to_load_downloads, e.message.toString()))
            } finally {
                binding.progressBar.visibility = View.GONE
                binding.swipeRefresh.isRefreshing = false
                isLoading = false
            }
        }

    }

    private fun loadMoreDownloads() {
        if (!hasMoreItems || isLoading) return
        loadLocalDownloads()
    }

    private fun loadThumbnailsForVisibleItems() {
        viewLifecycleOwner.lifecycleScope.launch {
            val layoutManager = binding.rvDownloads.layoutManager as LinearLayoutManager
            val firstVisible = layoutManager.findFirstVisibleItemPosition()
            val lastVisible = layoutManager.findLastVisibleItemPosition()

            if (firstVisible >= 0 && lastVisible >= firstVisible && firstVisible < adapter.itemCount) {
                val endIndex = (lastVisible + 1).coerceAtMost(adapter.itemCount)

                for (i in firstVisible until endIndex) {
                    val item = adapter.getItems()[i]
                    if (item.thumbnail == null) {
                        val thumbnail = LocalDownloadsScanner.loadThumbnailForItem(
                            requireContext(),
                            item
                        )
                        thumbnail?.let {
                            adapter.updateItemThumbnail(item.id, it)
                        }
                    }
                }
            }
        }
    }

    private fun loadThumbnailsForVisibleRange(firstVisible: Int, lastVisible: Int) {
        if (firstVisible !in 0..lastVisible) return

        for (i in firstVisible..lastVisible) {
            val item = adapter.getItemAt(i)
            item?.let {
                // Adapter will handle thumbnail loading automatically
                // By onBindViewHolder and loadThumbnailForVisibleItem
            }
        }
    }

    private fun showDeleteConfirmation(item: LocalDownloadItem) {
        AlertDialog.Builder(requireContext())
            .setTitle(getString(R.string.delete_file))
            .setMessage(getString(R.string.title_dialog_delete_file, item.displayName))
            .setPositiveButton(getString(R.string.delete_file)) { _, _ ->
                deleteFile(item)
            }
            .setNegativeButton(getString(R.string.cancel), null)
            .show()
    }

    private fun deleteFile(item: LocalDownloadItem) {
        viewLifecycleOwner.lifecycleScope.launch {
            try {
                val deleted = deleteFileFromStorage(item)

                if (deleted) {
                    adapter.removeItem(item.id)

                    requireContext().showToast(getString(R.string.delete_file_success))

                    if (adapter.itemCount == 0) {
                        binding.tvEmpty.visibility = View.VISIBLE
                        binding.rvDownloads.visibility = View.GONE
                    }
                } else {
                    requireContext().showToast(getString(R.string.delete_file_failed))
                }
            } catch (e: Exception) {
                e.printStackTrace()
                requireContext().showToast(getString(R.string.error, e.message))
            }
        }
    }

    private suspend fun deleteFileFromStorage(item: LocalDownloadItem): Boolean {
        return kotlin.runCatching {
            val rowsDeleted = requireContext().contentResolver.delete(
                item.uri,
                null,
                null
            )

            if (rowsDeleted > 0) return true

            item.filePath?.let { path ->
                val file = File(path)
                if (file.exists() && file.delete()) {
                    requireContext().contentResolver.delete(
                        item.uri,
                        null,
                        null
                    )
                    return true
                }
            }

            false
        }.getOrElse { false }
    }
	private fun hasRequiredReadPermissions(): Boolean {
        val result = if (android.os.Build.VERSION.SDK_INT >= 33) {
            val videoPerm = ContextCompat.checkSelfPermission(requireContext(), Manifest.permission.READ_MEDIA_VIDEO) == PackageManager.PERMISSION_GRANTED
            val audioPerm = ContextCompat.checkSelfPermission(requireContext(), Manifest.permission.READ_MEDIA_AUDIO) == PackageManager.PERMISSION_GRANTED
            println("DEBUG Permissions: READ_MEDIA_VIDEO: $videoPerm, READ_MEDIA_AUDIO: $audioPerm")
            videoPerm && audioPerm
        } else {
            val storagePerm = ContextCompat.checkSelfPermission(requireContext(), Manifest.permission.READ_EXTERNAL_STORAGE) == PackageManager.PERMISSION_GRANTED
            println("DEBUG Permissions: READ_EXTERNAL_STORAGE: $storagePerm")
            storagePerm
        }
        return result
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
        adapter.cancelAllThumbnailLoading()
        LocalDownloadsScanner.ThumbnailCache.clear()
        _binding = null
    }
}
