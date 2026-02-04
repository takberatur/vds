package com.agcforge.videodownloader.helper

import android.app.Activity
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.FrameLayout
import com.agcforge.videodownloader.R
import com.google.android.gms.ads.AdListener
import com.google.android.gms.ads.AdRequest
import com.google.android.gms.ads.AdView
import com.google.android.gms.ads.LoadAdError
import com.startapp.sdk.ads.banner.Banner
import com.startapp.sdk.ads.banner.BannerListener
import com.unity3d.services.banners.BannerErrorInfo
import com.unity3d.services.banners.BannerView
import com.unity3d.services.banners.UnityBannerSize

class BannerAdsHelper(private val activity: Activity) {

    private val TAG = "BannerAdsHelper"

    // Banner views
    private var admobBannerView: AdView? = null
    private var unityBannerView: BannerView? = null
    private var startIoBanner: Banner? = null

    // Containers
    private var admobContainer: View? = null
    private var unityContainer: FrameLayout? = null
    private var startIoContainer: FrameLayout? = null

    // Loading states
    private var isAdmobLoading = false
    private var isUnityLoading = false
    private var isStartIoLoading = false

    // Current provider
    private var currentProvider: AdsProvider? = null

    /**
     * Inflate banner view from layout
     */
    fun inflateBannerView(parent: ViewGroup): View {
        return LayoutInflater.from(activity)
            .inflate(R.layout.item_banner_ad, parent, false)
    }

    /**
     * Load and attach banner to container
     */
    fun loadAndAttachBanner(
        container: ViewGroup,
        onBannerLoaded: ((AdsProvider) -> Unit)? = null,
        onBannerFailed: ((AdsProvider) -> Unit)? = null
    ) {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled")
            container.visibility = View.GONE
            return
        }

        // Inflate banner layout
        val bannerView = inflateBannerView(container)
        container.addView(bannerView)

        // Get references
        admobContainer = bannerView.findViewById(R.id.admobBannerView)
        unityContainer = bannerView.findViewById(R.id.unityBannerContainer)
        startIoContainer = bannerView.findViewById(R.id.startIoBannerContainer)

        admobBannerView = admobContainer as? AdView

        // Load banner
        loadBanner(onBannerLoaded, onBannerFailed)
    }

    /**
     * Load banner with auto rotation
     */
    fun loadBanner(
        onBannerLoaded: ((AdsProvider) -> Unit)? = null,
        onBannerFailed: ((AdsProvider) -> Unit)? = null
    ) {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled")
            return
        }

        // Try providers in priority order
        for (provider in AdsConfig.Rotation.getBannerPriority()) {
            when (provider) {
                AdsProvider.ADMOB -> {
                    if (admobBannerView != null && !isAdmobLoading) {
                        loadAdmobBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
                AdsProvider.UNITY -> {
                    if (unityContainer != null && !isUnityLoading) {
                        loadUnityBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
                AdsProvider.STARTIO -> {
                    if (startIoContainer != null && !isStartIoLoading) {
                        loadStartIoBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
            }
        }
    }

    // ========== ADMOB BANNER ==========

    private fun loadAdmobBanner(
        onBannerLoaded: ((AdsProvider) -> Unit)?,
        onBannerFailed: ((AdsProvider) -> Unit)?
    ) {
        admobBannerView?.apply {
            isAdmobLoading = true

            adListener = object : AdListener() {
                override fun onAdLoaded() {
                    Log.d(TAG, "Admob banner loaded")
                    isAdmobLoading = false
                    currentProvider = AdsProvider.ADMOB
                    showBanner(AdsProvider.ADMOB)
                    onBannerLoaded?.invoke(AdsProvider.ADMOB)
                }

                override fun onAdFailedToLoad(error: LoadAdError) {
                    Log.e(TAG, "Admob banner failed: ${error.message}")
                    isAdmobLoading = false
                    tryNextProvider(AdsProvider.ADMOB, onBannerLoaded, onBannerFailed)
                }

                override fun onAdClicked() {
                    Log.d(TAG, "Admob banner clicked")
                }
            }

            loadAd(AdRequest.Builder().build())
        }
    }

    // ========== UNITY BANNER ==========

    private fun loadUnityBanner(
        onBannerLoaded: ((AdsProvider) -> Unit)?,
        onBannerFailed: ((AdsProvider) -> Unit)?
    ) {
        unityContainer?.apply {
            isUnityLoading = true
            removeAllViews()

            unityBannerView = BannerView(
                activity,
                AdsConfig.UNITY_BANNER_ID,
                UnityBannerSize(320, 50)
            ).apply {
                listener = object : BannerView.IListener {
                    override fun onBannerLoaded(bannerView: BannerView?) {
                        Log.d(TAG, "Unity banner loaded")
                        isUnityLoading = false
                        currentProvider = AdsProvider.UNITY
                        showBanner(AdsProvider.UNITY)
                        onBannerLoaded?.invoke(AdsProvider.UNITY)
                    }

                    override fun onBannerShown(bannerAdView: BannerView?) {}

                    override fun onBannerFailedToLoad(
                        bannerView: BannerView?,
                        errorInfo: BannerErrorInfo?
                    ) {
                        Log.e(TAG, "Unity banner failed: ${errorInfo?.errorMessage}")
                        isUnityLoading = false
                        tryNextProvider(AdsProvider.UNITY, onBannerLoaded, onBannerFailed)
                    }

                    override fun onBannerClick(bannerView: BannerView?) {
                        Log.d(TAG, "Unity banner clicked")
                    }

                    override fun onBannerLeftApplication(bannerView: BannerView?) {
                        Log.d(TAG, "Unity banner left application")
                    }
                }
                load()
            }

            addView(unityBannerView)
        }
    }

    // ========== START.IO BANNER ==========

    private fun loadStartIoBanner(
        onBannerLoaded: ((AdsProvider) -> Unit)?,
        onBannerFailed: ((AdsProvider) -> Unit)?
    ) {
        startIoContainer?.apply {
            isStartIoLoading = true
            removeAllViews()

            startIoBanner = Banner(activity, object : BannerListener {
                override fun onReceiveAd(view: View) {
                    Log.d(TAG, "Start.io banner loaded")
                    isStartIoLoading = false
                    currentProvider = AdsProvider.STARTIO
                    showBanner(AdsProvider.STARTIO)
                    onBannerLoaded?.invoke(AdsProvider.STARTIO)
                }

                override fun onFailedToReceiveAd(view: View?) {
                    Log.e(TAG, "Start.io banner failed to load.")
                    isStartIoLoading = false
                    tryNextProvider(AdsProvider.STARTIO, onBannerLoaded, onBannerFailed)
                }

                override fun onImpression(view: View) {}
                override fun onClick(view: View) {}
            })

            addView(startIoBanner)
        }
    }

    // ========== HELPER METHODS ==========

    private fun showBanner(provider: AdsProvider) {
        admobContainer?.visibility = View.GONE
        unityContainer?.visibility = View.GONE
        startIoContainer?.visibility = View.GONE

        when (provider) {
            AdsProvider.ADMOB -> admobContainer?.visibility = View.VISIBLE
            AdsProvider.UNITY -> unityContainer?.visibility = View.VISIBLE
            AdsProvider.STARTIO -> startIoContainer?.visibility = View.VISIBLE
        }
    }

    private fun tryNextProvider(
        failedProvider: AdsProvider,
        onBannerLoaded: ((AdsProvider) -> Unit)?,
        onBannerFailed: ((AdsProvider) -> Unit)?
    ) {
        val currentIndex = AdsConfig.Rotation.getBannerPriority().indexOf(failedProvider)
        val nextProviders = AdsConfig.Rotation.getBannerPriority().drop(currentIndex + 1)

        for (provider in nextProviders) {
            when (provider) {
                AdsProvider.ADMOB -> {
                    if (admobBannerView != null && !isAdmobLoading) {
                        loadAdmobBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
                AdsProvider.UNITY -> {
                    if (unityContainer != null && !isUnityLoading) {
                        loadUnityBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
                AdsProvider.STARTIO -> {
                    if (startIoContainer != null && !isStartIoLoading) {
                        loadStartIoBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
            }
        }

        Log.e(TAG, "All banner providers failed")
        hideBanner()
        onBannerFailed?.invoke(failedProvider)
    }

    fun hideBanner() {
        admobContainer?.visibility = View.GONE
        unityContainer?.visibility = View.GONE
        startIoContainer?.visibility = View.GONE
    }

    fun showCurrentBanner() {
        currentProvider?.let { showBanner(it) }
    }

    fun destroy() {
        admobBannerView?.destroy()
        unityBannerView?.destroy()
        startIoBanner = null

        admobBannerView = null
        unityBannerView = null
    }

    fun pause() {
        admobBannerView?.pause()
    }

    fun resume() {
        admobBannerView?.resume()
    }
}
